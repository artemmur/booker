import logging

from telethon import TelegramClient

from .app import CRM
from .bot import Bot

logging.basicConfig(
    format="[%(levelname) 5s/%(asctime)s] %(name)s: %(message)s", level=logging.INFO
)

Bot(
    TelegramClient("anon", "", ""),
    CRM(drive=None, sheet=None),
).start()
