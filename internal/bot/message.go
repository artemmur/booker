package bot

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/tg"
	"github.com/pkg/errors"
)

// MessageEvent represents message event.
type MessageEvent struct {
	Peer    tg.InputPeerClass
	Message *tg.Message

	user    *tg.User
	chat    *tg.Chat
	channel *tg.Channel

	baseEvent
}

// User returns User object and true if message got from user.
// False and nil otherwise.
func (e MessageEvent) User() (*tg.User, bool) {
	return e.user, e.user != nil
}

// Chat returns Chat object and true if message got from chat.
// False and nil otherwise.
func (e MessageEvent) Chat() (*tg.Chat, bool) {
	return e.chat, e.chat != nil
}

// Channel returns Channel object and true if message got from channel.
// False and nil otherwise.
func (e MessageEvent) Channel() (*tg.Channel, bool) {
	return e.channel, e.channel != nil
}

// WithReply calls given callback if current message event is a reply message.
func (e MessageEvent) WithReply(ctx context.Context, cb func(reply *tg.Message) error) error {
	h, ok := e.Message.GetReplyTo()
	if !ok {
		if _, err := e.Reply().Text(ctx, "Message must be a reply"); err != nil {
			return fmt.Errorf("%w: %v", err, "send")
		}
		return nil
	}

	var (
		msg *tg.Message
		err error
		log = e.logger.With(
			zap.Int("msg_id", e.Message.ID),
			zap.Int("reply_to_msg_id", h.ReplyToMsgID),
		)
	)
	switch p := e.Peer.(type) {
	case *tg.InputPeerChannel:
		log.Info("Fetching message", zap.Int64("channel_id", p.ChannelID))

		msg, err = e.getChannelMessage(ctx, &tg.InputChannel{
			ChannelID:  p.ChannelID,
			AccessHash: p.AccessHash,
		}, h.ReplyToMsgID)
	case *tg.InputPeerChat:
		log.Info("Fetching message", zap.Int64("chat_id", p.ChatID))

		msg, err = e.getMessage(ctx, h.ReplyToMsgID)
	case *tg.InputPeerUser:
		log.Info("Fetching message", zap.Int64("user_id", p.UserID))

		msg, err = e.getMessage(ctx, h.ReplyToMsgID)
	}
	if err != nil {
		log.Warn("Fetch message", zap.Error(err))
		if _, err := e.Reply().Textf(ctx, "Message %d not found", h.ReplyToMsgID); err != nil {
			return fmt.Errorf("%w: %v", err, "send")
		}
		return nil
	}

	return cb(msg)
}

// Reply creates new message builder to reply.
func (e MessageEvent) Reply() *message.Builder {
	return e.sender.To(e.Peer).ReplyMsg(e.Message)
}

// MessageHandler is a simple message event handler.
type MessageHandler interface {
	OnMessage(ctx context.Context, e MessageEvent) error
}

// MessageHandlerFunc is a functional adapter for Handler.
type MessageHandlerFunc func(ctx context.Context, e MessageEvent) error

// OnMessage implements MessageHandler.
func (h MessageHandlerFunc) OnMessage(ctx context.Context, e MessageEvent) error {
	return h(ctx, e)
}

func (b *Bot) handleUser(ctx context.Context, user *tg.User, m *tg.Message) error {
	b.logger.Info("Got message",
		zap.String("text", m.Message),
		zap.Int64("user_id", user.ID),
		zap.String("user_first_name", user.FirstName),
		zap.String("username", user.Username),
	)

	return b.onMessage.OnMessage(ctx, MessageEvent{
		Peer:      user.AsInputPeer(),
		user:      user,
		Message:   m,
		baseEvent: b.baseEvent(),
	})
}

func (b *Bot) handleChat(ctx context.Context, chat *tg.Chat, m *tg.Message) error {
	b.logger.Info("Got message from chat",
		zap.String("text", m.Message),
		zap.Int64("chat_id", chat.ID),
	)

	return b.onMessage.OnMessage(ctx, MessageEvent{
		Peer:      chat.AsInputPeer(),
		chat:      chat,
		Message:   m,
		baseEvent: b.baseEvent(),
	})
}

func (b *Bot) handleChannel(ctx context.Context, channel *tg.Channel, m *tg.Message) error {
	b.logger.Info("Got message from channel",
		zap.String("text", m.Message),
		zap.String("username", channel.Username),
		zap.Int64("channel_id", channel.ID),
	)

	return b.onMessage.OnMessage(ctx, MessageEvent{
		Peer:      channel.AsInputPeer(),
		channel:   channel,
		Message:   m,
		baseEvent: b.baseEvent(),
	})
}

func (b *Bot) handleMessage(ctx context.Context, e tg.Entities, msg tg.MessageClass) error {
	switch m := msg.(type) {
	case *tg.Message:
		if m.Out {
			return nil
		}

		switch p := m.PeerID.(type) {
		case *tg.PeerUser:
			user, ok := e.Users[p.UserID]
			if !ok {
				return errors.Errorf("unknown user ID %d", p.UserID)
			}
			return b.handleUser(ctx, user, m)
		case *tg.PeerChat:
			chat, ok := e.Chats[p.ChatID]
			if !ok {
				return errors.Errorf("unknown chat ID %d", p.ChatID)
			}
			return b.handleChat(ctx, chat, m)
		case *tg.PeerChannel:
			channel, ok := e.Channels[p.ChannelID]
			if !ok {
				return errors.Errorf("unknown channel ID %d", p.ChannelID)
			}
			return b.handleChannel(ctx, channel, m)
		}
	}

	return nil
}

func (b *Bot) OnNewMessage(ctx context.Context, e tg.Entities, u *tg.UpdateNewMessage) error {
	if err := b.handleMessage(ctx, e, u.Message); err != nil {
		if !tg.IsUserBlocked(err) {
			return errors.Wrapf(err, "handle message %d", u.Message.GetID())
		}

		b.logger.Debug("Bot is blocked by user")
	}
	return nil
}

func (b *Bot) OnNewChannelMessage(ctx context.Context, e tg.Entities, u *tg.UpdateNewChannelMessage) error {
	if err := b.handleMessage(ctx, e, u.Message); err != nil {
		return errors.Wrap(err, "handle")
	}
	return nil
}
