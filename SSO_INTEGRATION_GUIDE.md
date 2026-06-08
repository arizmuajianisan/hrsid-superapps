# SSO Integration Guide — HRSID

This guide explains how to integrate a third-party application into HRSID so that clicking its icon in the SSO portal automatically logs the user in — no second login required.

---

## Table of Contents

1. [How the Flow Works](#1-how-the-flow-works)
2. [What HRSID Provides](#2-what-hrsid-provides)
3. [Step-by-Step Integration](#3-step-by-step-integration)
4. [User Data Payload](#4-user-data-payload)
5. [Implementation Examples](#5-implementation-examples)
6. [Registering Your App in HRSID](#6-registering-your-app-in-hrsid)
7. [Security Notes](#7-security-notes)
8. [Error Reference](#8-error-reference)

---

## 1. How the Flow Works

The mechanism is a **One-Time Token (OTT)** exchange. The token is opaque (random, not guessable), single-use, and expires in **60 seconds**.

```
User (Browser)          SSO Portal (Frontend)      HRSID Backend         Your App Backend
     │                        │                          │                      │
     │  Click app icon        │                          │                      │
     │───────────────────────>│                          │                      │
     │                        │  POST /api/v1/launch/    │                      │
     │                        │  {app_slug}              │                      │
     │                        │  Authorization: Bearer   │                      │
     │                        │  <access_token>          │                      │
     │                        │─────────────────────────>│                      │
     │                        │                          │  Generate OTT        │
     │                        │                          │  Store in DB (60s)   │
     │                        │  { redirect_url }        │                      │
     │                        │<─────────────────────────│                      │
     │  Redirect browser to   │                          │                      │
     │  {base_url}/sso/       │                          │                      │
     │  callback?token=OTT    │                          │                      │
     │<───────────────────────│                          │                      │
     │                                                   │                      │
     │  GET /sso/callback?token=OTT                      │                      │
     │──────────────────────────────────────────────────────────────────────────>
     │                                                   │                      │
     │                                                   │  POST /api/v1/sso/   │
     │                                                   │  validate            │
     │                                                   │  { token: OTT }      │
     │                                                   │<─────────────────────│
     │                                                   │  Return user profile │
     │                                                   │─────────────────────>│
     │                                                   │                      │  Find/create
     │                                                   │                      │  local user
     │                                                   │                      │  Set session
     │  Redirect to app dashboard                        │                      │
     │<──────────────────────────────────────────────────────────────────────────
```

**Key properties:**

- The OTT is a **64-character random hex string** — impossible to guess
- It can only be used **once** (invalidated immediately after validation)
- It expires in **60 seconds** from creation
- Your app backend talks to HRSID backend **server-to-server** — the token never leaves the server after exchange

---

## 2. What HRSID Provides

### New endpoints (to be implemented in HRSID backend)

#### `POST /api/v1/launch/{app_slug}`

**Auth**: Required (Bearer token in Authorization header)

Creates a one-time token for launching the specified app. Verifies that the user's department has access to the app.

**Request**: No body required.

**Response `200 OK`**:

```json
{
  "redirect_url": "https://your-app.hirose.co.id/sso/callback?token=4a7f3c..."
}
```

**Response `403 Forbidden`** (user's department has no access):

```json
{ "error": "access denied" }
```

---

#### `POST /api/v1/sso/validate`

**Auth**: None (server-to-server call authenticated by `X-SSO-Secret` header)

Validates and consumes a one-time token. Returns the user's profile.

**Request headers**:

```
X-SSO-Secret: <shared_secret_configured_per_app>
Content-Type: application/json
```

**Request body**:

```json
{ "token": "4a7f3c..." }
```

**Response `200 OK`**:

```json
{
  "user": {
    "nik": "HE-001234",
    "email": "john.doe@hirose.co.id",
    "full_name": "John Doe",
    "role": "user",
    "department": "Engineering",
    "department_id": 1
  }
}
```

**Response `401 Unauthorized`** (token expired, already used, or invalid):

```json
{ "error": "invalid or expired token" }
```

---

## 3. Step-by-Step Integration

### Step 1 — Register your app in HRSID

An HRSID admin registers your app via the admin panel or directly in the database. See [Section 6](#6-registering-your-app-in-hrsid) for details. You will receive:

- Your app's **slug** (e.g., `hi-dsign`)
- A **shared secret** for validating tokens (`X-SSO-Secret` header)

---

### Step 2 — Add the SSO callback route to your app

Your app needs one new route: `GET /sso/callback`

This route receives the OTT from the URL query string and:

1. Calls HRSID to validate the token
2. Finds or creates the matching local user
3. Creates a local session / issues your app's own JWT
4. Redirects to the app dashboard

---

### Step 3 — Implement user matching

When HRSID returns the user profile, match it to your local `users` table:

```
HRSID NIK   ──→  your_users.nik   (preferred)
HRSID email ──→  your_users.email (fallback)
```

If no matching user exists, you have two options:

- **Auto-provision**: Create the local user from HRSID data (recommended for internal tools)
- **Block access**: Return `403` if the user is not pre-provisioned in your app

For **role mapping**, see the example in [Section 5](#5-implementation-examples).

---

### Step 4 — Protect the callback route

The callback route processes a token that has already been validated server-to-server — the browser only ever sends the opaque string. No sensitive data is exposed in the URL beyond the 60-second window.

However, you should still:

- Redirect with `303 See Other` after setting the session (avoid form resubmission)
- Set your session cookie as `HttpOnly`, `Secure`, and `SameSite=Lax`
- Never log or store the raw OTT value

---

## 4. User Data Payload

These are all the fields HRSID returns after a successful token validation:

| Field           | Type   | Description                                                |
| --------------- | ------ | ---------------------------------------------------------- |
| `nik`           | string | Employee ID (e.g., `HE-001234`) — unique, use for matching |
| `email`         | string | Corporate email — unique, use as fallback match            |
| `full_name`     | string | Display name                                               |
| `role`          | string | HRSID role: `"user"` or `"admin"`                          |
| `department`    | string | Department name (e.g., `"Engineering"`)                    |
| `department_id` | int    | Department internal ID                                     |

> **Important**: `role` here is the HRSID system role, not your app's role. Map it to your own role system as needed.

---

## 5. Implementation Examples

### Go (net/http)

```go
// SSO callback handler in your app
func (app *App) ssoCallbackHandler(w http.ResponseWriter, r *http.Request) {
    token := r.URL.Query().Get("token")
    if token == "" {
        http.Error(w, "missing token", http.StatusBadRequest)
        return
    }

    // Call HRSID to validate the token (server-to-server)
    user, err := validateSSOToken(token)
    if err != nil {
        http.Error(w, "invalid or expired SSO token", http.StatusUnauthorized)
        return
    }

    // Find or create local user by NIK
    localUser, err := app.db.FindOrCreateUserByNIK(user.NIK, user.Email, user.FullName)
    if err != nil {
        http.Error(w, "internal error", http.StatusInternalServerError)
        return
    }

    // Set your app's session cookie
    session := app.sessions.New(localUser.ID)
    http.SetCookie(w, &http.Cookie{
        Name:     "app_session",
        Value:    session.Token,
        HttpOnly: true,
        Secure:   true,
        SameSite: http.SameSiteLaxMode,
        Path:     "/",
    })

    http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

// HRSIDUser is the profile returned by HRSID
type HRSIDUser struct {
    NIK          string `json:"nik"`
    Email        string `json:"email"`
    FullName     string `json:"full_name"`
    Role         string `json:"role"`
    Department   string `json:"department"`
    DepartmentID int    `json:"department_id"`
}

func validateSSOToken(token string) (*HRSIDUser, error) {
    payload := map[string]string{"token": token}
    body, _ := json.Marshal(payload)

    req, _ := http.NewRequest("POST", "https://hrsid.hirose.co.id/api/v1/sso/validate", bytes.NewReader(body))
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("X-SSO-Secret", os.Getenv("HRSID_SSO_SECRET"))

    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("token validation failed: %d", resp.StatusCode)
    }

    var result struct {
        User HRSIDUser `json:"user"`
    }
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, err
    }
    return &result.User, nil
}
```

---

### Node.js / Express

```js
const express = require("express");
const axios = require("axios");

app.get("/sso/callback", async (req, res) => {
  const { token } = req.query;
  if (!token) return res.status(400).send("Missing token");

  let user;
  try {
    const { data } = await axios.post(
      "https://hrsid.hirose.co.id/api/v1/sso/validate",
      { token },
      { headers: { "X-SSO-Secret": process.env.HRSID_SSO_SECRET } },
    );
    user = data.user;
  } catch (err) {
    return res.status(401).send("Invalid or expired SSO token");
  }

  // Find or create local user
  let localUser = await db.users.findOne({ where: { nik: user.nik } });
  if (!localUser) {
    localUser = await db.users.create({
      nik: user.nik,
      email: user.email,
      full_name: user.full_name,
      // Map HRSID role to your app role:
      role: mapRole(user.role, user.department),
    });
  }

  req.session.userId = localUser.id;
  res.redirect("/dashboard");
});

function mapRole(hrsidRole, department) {
  if (hrsidRole === "admin") return "admin";
  if (department === "Engineering") return "engineer";
  return "viewer";
}
```

---

### PHP (Laravel)

```php
// routes/web.php
Route::get('/sso/callback', [SSOController::class, 'callback']);

// app/Http/Controllers/SSOController.php
class SSOController extends Controller
{
    public function callback(Request $request)
    {
        $token = $request->query('token');
        if (!$token) abort(400);

        $response = Http::withHeaders([
            'X-SSO-Secret' => config('services.hrsid.secret'),
        ])->post(config('services.hrsid.url') . '/api/v1/sso/validate', [
            'token' => $token,
        ]);

        if (!$response->ok()) abort(401);

        $hrsidUser = $response->json('user');

        // Find or create local user
        $user = User::updateOrCreate(
            ['nik' => $hrsidUser['nik']],
            [
                'email'     => $hrsidUser['email'],
                'full_name' => $hrsidUser['full_name'],
                'role'      => $this->mapRole($hrsidUser['role']),
            ]
        );

        Auth::login($user);
        return redirect('/dashboard');
    }

    private function mapRole(string $hrsidRole): string
    {
        return match($hrsidRole) {
            'admin' => 'admin',
            default => 'user',
        };
    }
}
```

---

### Python (FastAPI)

```python
import httpx
import os
from fastapi import FastAPI, HTTPException
from fastapi.responses import RedirectResponse

app = FastAPI()
HRSID_BASE_URL = os.getenv("HRSID_BASE_URL", "https://hrsid.hirose.co.id")
HRSID_SSO_SECRET = os.getenv("HRSID_SSO_SECRET")

@app.get("/sso/callback")
async def sso_callback(token: str):
    async with httpx.AsyncClient() as client:
        resp = await client.post(
            f"{HRSID_BASE_URL}/api/v1/sso/validate",
            json={"token": token},
            headers={"X-SSO-Secret": HRSID_SSO_SECRET},
        )

    if resp.status_code != 200:
        raise HTTPException(status_code=401, detail="Invalid or expired SSO token")

    hrsid_user = resp.json()["user"]

    # Find or create user in your DB
    local_user = await find_or_create_user(
        nik=hrsid_user["nik"],
        email=hrsid_user["email"],
        full_name=hrsid_user["full_name"],
    )

    # Set session / issue JWT — depends on your auth library
    session_token = create_session(local_user.id)

    response = RedirectResponse(url="/dashboard", status_code=303)
    response.set_cookie("session", session_token, httponly=True, secure=True, samesite="lax")
    return response
```

---

## 6. Registering Your App in HRSID

An HRSID admin registers your app in the Admin → Applications panel with the following fields:

| Field             | Example                         | Notes                                 |
| ----------------- | ------------------------------- | ------------------------------------- |
| Name              | `Hi-DSign`                      | Display name                          |
| Slug              | `hi-dsign`                      | URL-safe, unique identifier           |
| Base URL          | `https://hi-dsign.hirose.co.id` | Used to construct the redirect URL    |
| Icon URL          | `https://...`                   | Optional — shown in app launcher      |
| Description       | `Document signing app`          | Optional                              |
| Department Access | `Engineering`, `QA`             | Which departments can launch this app |

After registration, the admin generates and gives you a **shared secret** (`HRSID_SSO_SECRET`) to authenticate your backend's calls to `/api/v1/sso/validate`.

Store this secret in your app's environment variables — never commit it to source code.

---

## 7. Security Notes

| Concern                       | How it's handled                                                                            |
| ----------------------------- | ------------------------------------------------------------------------------------------- |
| Token guessing                | Token is 64 hex chars (256 bits) of `crypto/rand`                                           |
| Token replay                  | Token is invalidated immediately after first use                                            |
| Token expiry                  | Token expires after 60 seconds                                                              |
| Man-in-the-middle             | Use HTTPS for both HRSID and your app                                                       |
| Unauthorized validation calls | `X-SSO-Secret` shared secret per app                                                        |
| Department access bypass      | HRSID checks department access before issuing OTT                                           |
| Session fixation              | Generate a new session ID after SSO login in your app                                       |
| Open redirect                 | OTT redirect URL is fully constructed by HRSID — the browser never supplies the destination |

---

## 8. Error Reference

### From `POST /api/v1/launch/{app_slug}`

| HTTP | Error                   | Cause                                       |
| ---- | ----------------------- | ------------------------------------------- |
| 401  | `unauthorized`          | Missing or invalid access token             |
| 403  | `access denied`         | User's department has no access to this app |
| 404  | `application not found` | Unknown slug                                |
| 500  | `internal server error` | DB error                                    |

### From `POST /api/v1/sso/validate`

| HTTP | Error                      | Cause                                     |
| ---- | -------------------------- | ----------------------------------------- |
| 400  | `missing token`            | Empty token in body                       |
| 401  | `invalid or expired token` | Token not found, already used, or expired |
| 401  | `unauthorized`             | Missing or wrong `X-SSO-Secret` header    |
| 500  | `internal server error`    | DB error                                  |

---

## Summary Checklist for App Integrators

- [ ] App registered in HRSID admin panel with correct `base_url`
- [ ] Department access configured in HRSID for this app
- [ ] `HRSID_SSO_SECRET` added to your app's environment variables
- [ ] `GET /sso/callback` route implemented
- [ ] Token validated server-to-server (never browser-to-HRSID)
- [ ] Local user matched by `nik` (fallback: `email`)
- [ ] Session set as `HttpOnly`, `Secure`, `SameSite=Lax`
- [ ] Redirect to dashboard with `303 See Other`
- [ ] OTT never logged or stored by your app
