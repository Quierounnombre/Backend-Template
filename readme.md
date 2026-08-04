# Backend Template


A production-ready auth backend you can clone, configure, and deploy in about an hour. I would love to be around 15 minutes, but let's be real, just reading this would take about 10, so you have 50 minutes left.


Built around the assumption that an email is all you need for building a product and communicating with your user.

---

## License & Documentation

[![License](https://img.shields.io/badge/license-MIT-green)](LICENSE)

[![Docs](https://img.shields.io/badge/docs-quierounnombre.github.io-blue)](https://quierounnombre.github.io/api/janus_api)

The upgraded version (multi-tenancy, admin UI, audit logging) is closed-source and available separately. Reach out on LinkedIn if interested.

---


## What's included


| Area | Details |
|---|---|
| **Auth** | OAuth2 (Google) + password login, users can use either |
| **JWT** | RS256 asymmetric keys, validates and offers JWKS, scales horizontaly and resistant to tampering |
| **2FA** | Email OTP on registration, keeps bots out without adding friction for real users |
| **Passwords** | bcrypt hashing, reset flow included |
| **Database** | `pgxpool` connection pooling out of the box |
| **Rate limiting** | Token Bucket approach per IP + Path |
| **Email** | SMTP configured, worker + manager aproach for efficient use of resources. |
| **Config** | config driven via `config.yaml` for non-sensitive settings, `.env` for secrets |
| **Infra** | Dockerfile + Docker Compose |
| **Logging** | Log rotations, storage, compresion included |
| **Templates** | Email templates for customizing your brand |


> **On rate limiting at scale:** IP-based limiting will eventually lead to false-positives on shared IPs (offices, proxies). Fine for an MVP, and even <100 users/day. plan to swap it out before you're big enough to care.


---


## Security decisions worth knowing


- Unauthenticated endpoints don't reveal whether a user exists. This is intentional.
- Secrets never touch `config.yaml`. If you're committing your `.env` to a production repo, that's on you.
- OAuth and password auth share the same user model, no duplicates, if no password is set, they will need to hit password_reset.
- JWT uses an RS256 asymmetric keys, obtain the public key in JWKS format at `your_domain/auth/public-key`
- JWT has no revocation, for this i recomend at the moment to set a `15 mins` time between token refresh(PR to fix it are accepted)
- Your frontend needs a `reset_password_new` GET endpoint, that sends a POST to the `reset_password_new` (We don't want to leak user passwords, don't we?).


---

## Shared User Model

Users can log either through the `OAuth_Login` endpoint, or through the `Pass_Login` endpoint.

If the user comes first from `Pass_Signup` they will be able to use OAuth from the get go(they go through the 2FA, so the email they provide == theirs)

If the user comes from the `OAuth_Login` their password will be set to blank, and always rejected by `Pass_Signup` they will need to reset their password, once reseted, they will be able to log using password also.

---


## Launching the service(The last 15 minutes)


```bash
git clone git@github.com:Quierounnombre/Backend-Template.git
cp .env.example .env   # fill in your secrets
docker compose up --build
```

Configure `config.yaml` for everything else. It's commented, read it.

Lastly set up your templates in the `Template/` folder

**IN CASE OF DOUBT**: Use the defined values in the template.

### Validating Users from your service

Do something like this in your service

```go
jwks, _ := keyfunc.NewDefaultCtx(ctx, []string{"https://yourserver.com/auth/public-key"})
token, err := jwt.Parse(tokenString, jwks.Keyfunc, jwt.WithValidMethods([]string{"RS256"}))
if err != nil || !token.Valid {
	// reject
}
```

Then just set your email templates in the templates folder, and you are good to go!

---


## Stack


Go · Gin · pgx · Docker · RS256 JWT · Google OAuth2


---


## When to use this


You're starting a new backend, and:
- You need auth wired up fast, without debugging JWT or OAuth callback flows at 2am.
- Avoid regulatory risk and operational complexity.

For this, is simple, clone this, and just verify the tokens with the public key at your other services.


When **not** to use this: 
- If you don't need user accounts, this is overkill.
- You are at huge scales, many small decisions are made on the assumptions that you will have medium user counts.
- You like dolphins

## Need help?

Ask me on [LinkedIn](https://www.linkedin.com/in/vicente-garcia-andrade/), happy to talk :)
