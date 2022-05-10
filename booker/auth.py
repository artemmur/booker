import os

from google.auth.transport.requests import Request
from google.oauth2.credentials import Credentials
from google_auth_oauthlib.flow import InstalledAppFlow

CREAD_FILENAME = "credentials.json"
SCOPES = ["https://www.googleapis.com/auth/spreadsheets.readonly"]


def credentials(filename: str) -> Credentials:
    """Returns credentials for auth in google API"""
    creds: Credentials | None = None
    # The file token.json stores the user's access and refresh tokens, and is
    # created automatically when the authorization flow completes for the first
    # time.
    if os.path.exists(filename):
        creds = Credentials.from_authorized_user_file(filename, SCOPES)
    # If there are no (valid) credentials available, let the user log in.
    if not creds or not creds.valid:
        if creds and creds.expired and creds.refresh_token:
            creds.refresh(Request())
        else:
            flow: InstalledAppFlow = InstalledAppFlow.from_client_secrets_file(
                CREAD_FILENAME, SCOPES
            )
            creds = flow.run_local_server(port=0)
        # Save the credentials for the next run
        with open(filename, "w") as token:
            token.write(creds.to_json())
    return creds
