# 📊 CampaignHub Database Design

This repository contains the database schema design for **CampaignHub**, a platform that connects **brands** and **users** for running and participating in marketing campaigns.  
It supports transactions, campaign management, submissions, and support ticketing.

---

## 📐 ER Diagram

![Database Diagram](./database/CamapaignHub.png)  
*(Exported from [dbdiagram.io](https://dbdiagram.io))*  

---

## 🗄️ Database Schema

The schema is implemented in **PostgreSQL** and includes the following major entities:

### Core Tables
- **Users**: Stores creator/influencer details.
- **Brands**: Represents companies running campaigns.
- **Accounts**: Wallet/accounts for both users and brands.
- **Transactions**: Money flow between accounts.
- **Campaigns**: Marketing campaigns created by brands.
- **Submissions**: Content submitted by users for campaigns.
- **Platform Links**: Social profiles linked to users.

### Supporting Tables
- **Roles**: Defines user roles.
- **Status**: Tracks campaign/submission states.
- **Tx Status**: Transaction result (success/failed).
- **Ticket Status**: For support ticket workflow.
- **Support Tickets**: Customer support requests.

---

## 🔑 Key Features

- ✅ **Composite Keys**: e.g. (`userid`, `name`) in `platform_links`  
- ✅ **Indexes**: for fast queries (`campaign_id`, `status`, etc.)  
- ✅ **Constraints**: uniqueness, checks (age, budget, balances ≥ 0)  
- ✅ **Polymorphic Accounts**: `holder_type` allows `user` or `brand`  

---

## ⚙️ Setup Instructions

1. Clone this repository  
   ```bash
   git clone https://github.com/Alter-Sitanshu/campaignHub
   cd campaignHub
