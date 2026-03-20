# Google ID Setup Guide

This guide explains how to set up Google OAuth 2.0 and obtain a **Google Client ID** for the Titik Nol backend.

## 1. Create a Project in Google Cloud Console

1.  Open the [Google Cloud Console](https://console.cloud.google.com/).
2.  In the top-left corner, click the project dropdown and select **New Project**.
3.  Enter a project name (e.g., `titik-nol`) and click **Create**.

## 2. Configure OAuth Consent Screen

If this is your first time setting up OAuth in this project, you must configure the consent screen:

1.  Go to **APIs & Services** > **OAuth consent screen** in the sidebar.
2.  Select **User Type**:
    *   **External**: Available to any user with a Google Account.
    *   **Internal**: Only available to users within your organization (if using Google Workspace).
3.  Click **Create**.
4.  Fill in the required information:
    *   **App name**: `Titik Nol`
    *   **User support email**: Your email.
    *   **Developer contact information**: Your email.
5.  Click **Save and Continue** until you return to the dashboard. (You can skip "Scopes" and "Test Users" for now, but adding test users is recommended if the app is in 'Testing' mode).

## 3. Create OAuth 2.0 Client ID

1.  Go to **APIs & Services** > **Credentials**.
2.  Click **+ Create Credentials** at the top and select **OAuth client ID**.
3.  Choose **Application type**:
    *   For testing with common frontends (e.g., React, Vue, Next.js), select **Web application**.
4.  Enter a name (e.g., `Titik Nol Web Client`).
5.  **Authorized JavaScript origins**:
    *   Add your frontend URL (e.g., `http://localhost:3000`).
6.  **Authorized redirect URIs**:
    *   If you are using a library like `react-google-login` or the new Google Identity Services, you might not need a redirect URI on the backend, as the ID Token is often obtained on the frontend.
7.  Click **Create**.
8.  A dialog will appear showing your **Client ID** and **Client Secret**.

## 4. Configuration

1.  Copy the **Client ID**.
2.  Open your `.env` file in the project root.
3.  Paste the Client ID into the `GOOGLE_CLIENT_ID` variable:

```env
GOOGLE_CLIENT_ID=your-client-id-goes-here.apps.googleusercontent.com
```

## 5. Verification

The backend uses this Client ID to verify ID Tokens sent by the frontend in the Login/Register flow. Ensure the Client ID in `.env` matches the one used by your frontend application.
