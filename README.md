# JustJio 🎉 ![Endpoint Badge](https://img.shields.io/endpoint?url=https%3A%2F%2Fgomon.rowentey.xyz%2Fapi%2Fwebsites%2Fbadge%3FwebsiteUrl%3Dhttps%3A%2F%2Fjustjio-staging.rowentey.xyz) [![CI Pipeline](https://github.com/RowenTey/justjio/actions/workflows/ci.yml/badge.svg)](https://github.com/RowenTey/justjio/actions/workflows/ci.yml) [![Staging Environment](https://github.com/RowenTey/justjio/actions/workflows/staging_cd.yml/badge.svg)](https://github.com/RowenTey/justjio/actions/workflows/staging_cd.yml)

> A party-planning app that streamlines all the pain of hosting one 🍻

![landing](./client/public/assets/JustJio.gif)

## ⭐ Features

- **User Authentication & Management**

  - Login/Signu
  - OTP Verification
  - Profile Management

- **Room Management**

  - Create Rooms
  - Invite Friends
  - Room Types (Public and Private)
  - Room Actions (Join, Leave or Close)

- **Bill Splitting & Transactions**

  - Create Bills
  - Consolidate Bills
  - Transaction Tracking

- **Social Features**

  - Friends System
  - Search Users
  - Notifications

- **Messaging**

  - Real-time Chat
  - Message History

- **Push Notifications**

  - Web Push API

## 🛠 Getting Started

> See specific instructions from respective directories

## 📂 Project Folder Structure

### Top Level Directory Layout

```terminal
.
├── .github/                    # CI/CD pipeline
├── client/                     # react web app
├── server/
  ├── api/                      # go API server
  └── ws/                       # go websockets server
└── infra/                      # infra configs
```

## Planned Tasks (BE)

- Add validation layer
- Migrate to Inbox based concept in DB for unread messages paired with Redis streams for scaling Websocket server
- Sliding window counter rate-limiting
- Refactor integration tests with shared containers
- Add read-through cache

<!--

## 🧠 Contributors - Team OneStart 🏆🤟🏼

- [@RowenTey](https://github.com/RowenTey)
- [@czhi-heng](https://github.com/czhi-heng)
- [@JULU909](https://github.com/JULU909)
- [@Eldrick7](https://github.com/Eldrick7)
- [@cplAloysius](https://github.com/cplAloysius)
- [@amabellim](https://github.com/amabellim)

-->
