"""Describe business logic of booker crm application"""
import typing
from dataclasses import dataclass, field

from .drive import Drive
from .sheet import Sheet


@dataclass
class Request:
    """Model describes user appeal to compensation or payment"""

    username: str = str()
    amount: int = 0
    purpose: str = str()
    account: str = str()
    images: list[typing.Any] = field(default_factory=lambda: list())


@dataclass
class Response:
    """Model describe user appeal response"""

    username: str = str()
    amount: int = 0
    purpose: str = str()
    account: str = str()
    image_url: list[str] = field(default_factory=lambda: list())


class CRM:
    """Business logic of accounting CRM application"""

    def __init__(self, drive: Drive, sheet: Sheet) -> None:
        self.drive: Drive = drive
        self.sheet: Sheet = sheet

    def compensation(self, req: Request) -> Response:
        """New compensation for employee. Returns information about created record"""
        return Response(
            username=req.username,
            amount=req.amount,
            purpose=req.purpose,
            account=req.account,
        )

    def payment(self, req: Request) -> Response:
        """Creates new payment. Returns information about created record"""
        pass
