# Google ID Setup Guide

This guide explains how to set up Google OAuth 2.0 and obtain a **Google Client ID** for the Titik Nol backend. The backend uses this ID to verify ID Tokens sent by the frontend via Google Identity Services (GSI).

## 1. Create a Project in Google Cloud Console

1.  Open the [Google Cloud Console](https://console.cloud.google.com/).
2.  In the top-left corner, click the project dropdown and select **New Project**.
3.  Enter a project name (e.g., `titik-nol`) and click **Create**.
4.  Ensure your new project is selected in the top-left dropdown.

## 2. Configure OAuth Consent Screen

You must configure the consent screen before creating credentials:

1.  Go to **APIs & Services** > **OAuth consent screen** in the sidebar.
2.  Select **User Type**:
    *   **External**: Available to any user with a Google Account. (Recommended for development/production).
3.  Click **Create**.
4.  **App Information**:
    *   **App name**: `Titik Nol`
    *   **User support email**: Your email.
    *   **Developer contact information**: Your email.
5.  Click **Save and Continue**.
6.  **Scopes**: Click **Add or Remove Scopes**. Select:
    *   `.../auth/userinfo.email`
    *   `.../auth/userinfo.profile`
    *   `openid`
7.  Click **Save and Continue**.
8.  **Test Users** (CRITICAL):
    *   If your "Publishing status" is **Testing**, only accounts added here can log in.
    *   Click **+ Add Users** and add your own Google email address.
9.  Click **Save and Continue**, then click **Back to Dashboard**.

## 3. Create OAuth 2.0 Client ID

1.  Go to **APIs & Services** > **Credentials**.
2.  Click **+ Create Credentials** at the top and select **OAuth client ID**.
3.  Choose **Application type**: **Web application**.
4.  Enter a name (e.g., `Titik Nol Web Client`).
5.  **Authorized JavaScript origins**:
    *   Add your frontend development URL (e.g., `http://localhost:3000`).
    *   Add your production frontend URL when ready.
6.  **Authorized redirect URIs**:
    *   If using **Google Identity Services (GSI)** with the popup login (common for SPAs), you may leave this empty.
    *   If you use a redirect-based flow, add your frontend redirect URL.
7.  Click **Create**.
8.  Copy the **Client ID** from the dialog that appears. (You don't need the Client Secret for ID Token verification on the backend).

## 4. Backend Configuration

1.  Open your `.env` file in the project root.
2.  Paste the Client ID into the `GOOGLE_CLIENT_ID` variable:

```env
GOOGLE_CLIENT_ID=your-client-id-goes-here.apps.googleusercontent.com
```

3.  Restart the backend server for changes to take effect.

## 5. Common Pitfalls

*   **Missing Test User**: If you get a "403 Access Denied" or "Unauthorized" error even with the correct Client ID, ensure your email is added to the **Test Users** section in the OAuth consent screen.
*   **Redirect URI / Origin Mismatch**: Ensure the URL in "Authorized JavaScript origins" exactly matches the URL you're using to access the frontend (including `http://` vs `https://` and ports).
*   **Token Verification Error**: If the backend says the token is invalid, verify that the `GOOGLE_CLIENT_ID` in `.env` is exactly the same one used by the frontend to request the token.
