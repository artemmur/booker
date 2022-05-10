"""Implementation of chat bot interface"""
import dataclasses
import logging
import typing

from telethon import Button, TelegramClient, events

from .app import CRM, Request, Response
from .tmpl import (
    CHOOSE,
    COMPENSATION,
    COMPENSATION_ACCOUNT,
    COMPENSATION_AMOUNT,
    COMPENSATION_CREATED,
    COMPENSATION_PURPOSE,
    GREETING,
    INVALID_INPUT,
    PING,
    USAGE,
)


class Bot:
    """Bot is a wrapper for telethon library"""

    def __init__(self, client: TelegramClient, app: CRM) -> None:
        self._client: TelegramClient = client
        self._app: CRM = app

    def start(self) -> None:
        """Start bot listens events"""
        # Register all handlers.
        for h in [self.ping, self.appeal, self.usage, self.greeting]:
            self._client.add_event_handler(h)

        # Set message parse mode to html
        self._client.parse_mode = "html"

        logging.info("starting chat bot...")
        self._client.start()
        self._client.run_until_disconnected()

    @events.register(events.NewMessage(pattern="/appeal"))
    async def appeal(self, event: events.NewMessage.Event) -> None:
        """Create new appeal from user to account office"""
        sender: typing.Any = await event.get_sender()
        logging.info(f"new request by {sender.username}")

        # Start conversation with user
        async with self._client.conversation(event.chat_id) as conv:
            # Ask user about action
            await conv.send_message(
                CHOOSE,
                buttons=[
                    Button.inline("compensation"),
                    Button.inline("payment"),
                ],
            )

            # Handle user input
            resp: typing.Any = await conv.wait_event(
                events.CallbackQuery(event.chat_id)
            )

            match resp.data:
                case b"compensation":
                    req: Request = Request()
                    req.username = sender.username

                    await conv.send_message(COMPENSATION_PURPOSE)
                    purpose: typing.Any = await conv.get_response()
                    req.purpose = purpose.text

                    await conv.send_message(COMPENSATION_AMOUNT)
                    amount: typing.Any = await conv.get_response()
                    try:
                        req.amount = int(amount.text)
                    except TypeError as err:
                        await amount.reply(INVALID_INPUT)
                        logging.error(f"invalid input by {sender.username}: {err}")
                        return

                    await conv.send_message(COMPENSATION_ACCOUNT)
                    account: typing.Any = await conv.get_response()
                    req.account = account.text

                    resp: Response = self._app.compensation(req)
                    await conv.send_message(COMPENSATION_CREATED)
                    await conv.send_message(
                        COMPENSATION.format_map(dataclasses.asdict(resp))
                    )
                    return

                case b"payment":
                    pass

                case _ as invalid:
                    await conv.send_message(INVALID_INPUT)
                    logging.error(f"invalid input by {sender.username}: {invalid}")
                    return

    @events.register(events.NewMessage(pattern="/ping"))
    async def ping(self, event: events.NewMessage.Event) -> None:
        """Bot healthcheck handler"""
        await event.reply(PING)

    @events.register(events.NewMessage(pattern="/start"))
    async def greeting(self, event: events.NewMessage.Event) -> None:
        """Greeting new user"""
        sender: typing.Any = await event.get_sender()
        await self._client.send_message(
            event.chat_id,
            GREETING.format(user=sender.username),
        )
        await self.usage(event)

    @events.register(events.NewMessage(pattern="/help"))
    async def usage(self, event: events.NewMessage.Event) -> None:
        """Usage information"""
        await self._client.send_message(event.chat_id, USAGE)
