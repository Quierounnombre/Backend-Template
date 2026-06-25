# Backend Template


A production-ready auth backend you can clone, configure, and deploy in about an hour. I would love to be around 15 minutes, but let's be real, just reading this would take about 10, so you have 50 minutes left.


Built around the assumption that an email is all you need for building a product and communicating with your user.


---


## What's included


| Area | Details |
|---|---|
| **Auth** | OAuth2 (Google) + password login, users can use either |
| **JWT** | RS256 asymmetric keys, scales horizontaly and resistant to tampering |
| **2FA** | Email OTP on registration, keeps bots out without adding friction for real users |
| **Passwords** | bcrypt hashing, reset flow included |
| **Database** | `pgxpool` connection pooling out of the box |
| **Rate limiting** | Fixed window per IP, good enough until you're at scale, MUST be revised once scaled |
| **Email** | Sending wired up. Works, but MUST be revised once scaled |
| **Config** | `config.yaml` for non-sensitive settings, `.env` for secrets |
| **Infra** | Dockerfile + Docker Compose |


> **On rate limiting at scale:** IP-based limiting will eventually lead to false-positives on shared IPs (offices, proxies). Fine for an MVP, and even <100 users/day. plan to swap it out before you're big enough to care.


---


## Security decisions worth knowing


- Unauthenticated endpoints don't reveal whether a user exists. This is intentional.
- Secrets never touch `config.yaml`. If you're committing your `.env` to a production repo, that's on you.
- OAuth and password auth share the same user model, no duplicates.
- JWT uses an RS256 asymmetric keys, obtain the public key at `your_domain/auth/public-key`


---


## Getting started


```bash
git clone <repo>
cp .env.example .env   # fill in your secrets
docker compose up
```


Configure `config.yaml` for everything else. It's commented, read it.


---


## Stack


Go · Gin · pgx · Docker · RS256 JWT · Google OAuth2


---


## When to use this


You're starting a new backend, you need auth wired up fast, without debugging JWT, OAuth callback flows at 2am. Clone this, and just verify the tokens with the public key at your other services.


When **not** to use this: 
- If you don't need user accounts, this is overkill.
- You are at huge scales, many small decisions are made on the assumptions are based on low user counts, this may require tuning in **MAIL** & **RATE LIMITER**
- You like dolphins
