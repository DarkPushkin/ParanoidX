# Evolution Plan: simplex-node → Sovereign Digital Nation

**Goal:** 10,000 users, $500K ARR, $10M+ valuation in 18 months

---

## 1. Current Situation (b116, June 2026)

| Metric | Value |
|--------|-------|
| Users | 1 (single admin node) |
| Subscriptions | $0/month |
| Investment | Self-funded |
| Valuation | $750K - $3M |
| Active contacts | ~15 (test) |
| Silver reserve | 53,704,000,000,000 ng |
| Backing ratio | 137,921x |
| Codebase | ~60,560 LOC, 100+ API endpoints |
| Tech readiness | TRL 6-7 (prototype → beta) |

### Strengths
- Complete sovereign node: messaging + economy + radio + AI + identity
- Silver-backed digital currency (Liquid Taler)
- Tor-native, censorship-resistant
- ParanoidX enterprise privacy suite
- 3 Telegram bots + AI Steward

### Weaknesses
- Zero users (chicken-and-egg problem)
- No mobile app on App Stores
- No fiat on-ramp
- No marketing / brand awareness
- Single developer dependency
- No legal entity

---

## 2. Target User Personas

### Persona A: Privacy-Conscious Professional (B2C)
- **Profile:** Journalist, activist, crypto-trader, remote worker
- **Pain:** WhatsApp/Telegram are not private enough
- **WTP:** $5-15/month for truly private comms + digital economy
- **Size:** ~500M globally

### Persona B: Enterprise / DAO / Community (B2B)
- **Profile:** DAO treasury manager, community operator, privacy team
- **Pain:** Need sovereign communication + treasury + governance in one
- **WTP:** $100-1000/month for node-as-a-service
- **Size:** ~50K DAOs + privacy teams globally

### Persona C: Island / Sovereign Individual (B2C Premium)
- **Profile:** "Digital nation" adopter, liberty-minded, crypto-native
- **Pain:** No platform respects true sovereignty
- **WTP:** $20-50/month for full stack membership
- **Size:** ~5M globally

---

## 3. Evolution Phases

### Phase 1: Foundation (Months 1-3) — Target: 100 users, $5K MRR

#### Technical
- [ ] Mobile SDK (iOS/Android) wrappers around existing API
- [ ] PWA web client as stop-gap (no `dart:io` dependency)
- [ ] Fiat on-ramp via Stripe/LemonSqueezy (buy ng with card)
- [ ] Referral system (invite code → free month)
- [ ] Onboarding wizard (guided tour of all features)

#### Growth
- [ ] Launch on SimpleX community channels (existing user base)
- [ ] Post to Telegram privacy groups, DAO forums, crypto subreddits
- [ ] Create 60-second demo video: "Your sovereign node in 1 minute"
- [ ] Privacy-focused Twitter/X account + build in public
- [ ] Offer 100 lifetime memberships at $100 pre-sale

#### Monetization
- **Free tier:** 1 contact, 100 msgs/day, no economy features
- **Pro tier ($10/mo):** Unlimited contacts, full economy, AI steward, 10GB vault
- **Business tier ($100/mo):** White-label node, 10 seats, API access, priority support

#### KPIs
| Metric | Target |
|--------|--------|
| Registered users | 100 |
| MRR | $5,000 |
| Conversion rate | 5% |
| Valuation | $5M |

#### Investment needed: $50,000
- Mobile SDK development: $20K
- Marketing (ads, content): $15K
- Server infrastructure: $10K
- Legal entity formation: $5K

---

### Phase 2: Product-Market Fit (Months 4-9) — Target: 1,000 users, $50K MRR

#### Technical
- [ ] iOS App Store launch (Swift wrapper)
- [ ] Google Play launch (Kotlin wrapper)
- [ ] Web client (WASM Flutter build)
- [ ] Push notifications (APNS + FCM)
- [ ] On-chain proof-of-reserve (TON)
- [ ] TON-based liquid staking derivative for ng
- [ ] Invite-only → open registration

#### Growth
- [ ] Product Hunt launch
- [ ] Partnerships with 5 privacy/liberty organizations
- [ ] Ambassador program (10% commission on referrals)
- [ ] Blog: "Week in review" → case studies → thought leadership
- [ ] Cold outreach to 100 DAOs with node proposal
- [ ] Press release: "World's First Silver-Backed Sovereign Network"

#### Monetization
- **Free tier** (retain, with limits)
- **Pro tier ($15/mo):** +Scheduled sends, auto-reply, templates, analytics
- **Business tier ($200/mo):** +Custom domain, webhook integration, dedicated bridge
- **Enterprise tier ($1K/mo):** +On-premise node, SLA, custom integration

#### KPIs
| Metric | Target |
|--------|--------|
| Registered users | 1,000 |
| MRR | $50,000 |
| DAU/MAU | 20% |
| NPS | >40 |
| Valuation | $10-15M |

#### Investment needed: $200,000
- Mobile app development: $80K
- Marketing & growth: $60K
- Infrastructure scaling: $30K
- Legal & compliance: $20K
- Security audit: $10K

---

### Phase 3: Scale (Months 10-18) — Target: 10,000 users, $500K MRR

#### Technical
- [ ] Decentralized autonomous treasury (DAO treasury module)
- [ ] Cross-chain atomic swaps (BTC, ETH, TON)
- [ ] AI-governed monetary policy
- [ ] ParanoidX Pro (enterprise privacy suite)
- [ ] RWA tokenization marketplace
- [ ] SDK for 3rd party integrations (REST + WebSocket)
- [ ] Horizontal scaling (multi-node federation)

#### Growth
- [ ] Affiliate network (500+ ambassadors)
- [ ] Referral contest with silver-backed rewards
- [ ] Localization (FR, ES, ZH, AR)
- [ ] B2B sales team (3 enterprise reps)
- [ ] Webinars: "Build your own digital nation"
- [ ] Conference presence (Bitcoin202x, EthCC, FOSDEM)

#### Monetization
- **Free tier** (user acquisition engine)
- **Pro tier ($20/mo):** +ParanoidX basics, labels, groups
- **Business tier ($300/mo):** +Advanced ParanoidX, dedicated onion, analytics
- **Enterprise tier ($2K/mo):** +On-prem, custom features, training
- **Transaction fees:** 0.5% on ng transfers
- **Channel subscriptions:** Content monetization on SimpleX channels

#### KPIs
| Metric | Target |
|--------|--------|
| Registered users | 10,000 |
| MRR | $500K |
| Annual run rate | $6M |
| Gross margin | 70% |
| Valuation | $50-100M |

#### Investment needed: $2,000,000
- Engineering team (5 senior engineers): $600K/yr
- Sales & marketing: $500K/yr
- Infrastructure: $300K/yr
- Legal & compliance: $200K/yr
- Security & audits: $200K/yr
- Operations & overhead: $200K/yr

---

## 4. Subscription & Revenue Model

### Tiered Pricing Strategy

```
                    Free          Pro           Business      Enterprise
Monthly             $0            $20           $300          $2,000
Yearly              $0            $200          $3,000        $20,000
Contacts            1             ∞             ∞             ∞
Economy             ✗             Full          Full          Full
Vault               ✗             10GB          100GB         ∞
AI Steward          ✗             ✓             ✓             ✓
ParanoidX           ✗             Light         Full          Full
Bridge              Shared        Shared        Dedicated     On-prem
API                 ✗             ✓             ✓             ✓
Support             Community     Email         Priority      24/7
Custom domain       ✗             ✗             ✓             ✓
SLA                 ✗             ✗             99.9%         99.99%
```

### Revenue Projection

| Tier | Phase 1 (Mo 3) | Phase 2 (Mo 9) | Phase 3 (Mo 18) |
|------|----------------|----------------|-----------------|
| Free | 50 users | 500 users | 5,000 users |
| Pro | 20 @ $20 | 300 @ $20 | 3,000 @ $20 |
| Business | 5 @ $300 | 50 @ $300 | 400 @ $300 |
| Enterprise | 1 @ $2K | 10 @ $2K | 50 @ $2K |
| Transaction fees | $0 | $500 | $10,000 |
| **MRR** | **$5,900** | **$57,500** | **$485,000** |

### Additional Revenue Streams
- **Silver-backed asset minting fees:** 0.1% of mint value
- **RWA registration fees:** $50 per asset
- **Channel subscriptions (creator economy):** 10% platform fee
- **SMP relay as-a-service:** $5/node/month
- **Consulting & custom deployments:** $10-50K per project
- **ParanoidX On-Prem license:** $5K/year

---

## 5. Investment Strategy

### Pre-Seed Round (Now — Month 3)
**Amount:** $100,000
**Source:** Self + angel investors (privacy/liberty-minded)
**Use:** Mobile SDK, marketing, legal entity
**Valuation:** $5M cap (SAFE note)
**Target:** 100 users, $5K MRR

### Seed Round (Month 4-9)
**Amount:** $1,000,000
**Source:** Crypto VCs (a16z, Polychain, Blockchain Capital), privacy funds
**Use:** Mobile launch, growth team, infrastructure
**Valuation:** $10-15M
**Target:** 1,000 users, $50K MRR
**Key pitch points:**
- First silver-backed sovereign network
- Complete stack (chat + economy + AI + privacy) — no competitor
- 100+ API endpoints, TRL 7
- ParanoidX as enterprise moat

### Series A (Month 10-18)
**Amount:** $5,000,000
**Source:** Institutional VCs, strategic investors, sovereign wealth funds
**Use:** Scaling engineering, global marketing, compliance
**Valuation:** $50-100M
**Target:** 10,000 users, $500K MRR
**Key pitch points:**
- Traction: 10K users, $6M ARR
- Proven unit economics
- Network effects (P2P relay, channel ecosystem)
- Government/corporate interest in sovereign comms

### Valuation Timeline

```
$100M  ┤                                                      ╱
       │                                                    ╱
$50M   ┤                                                ╱
       │                                              ╱
$15M   ┤                                          ╱
       │                                        ╱
$5M    ┤                                    ╱
       │                                  ╱
$1M    ┤  ╱─── Pre-seed                  ╱
       │╱              ╱────────── Series A
       │  ╱── Seed ───╱
       │ ╱
       ├──────┬──────┬──────┬──────┬──────┬──────┬──────┬
      Jun'26 Sep'26 Dec'26 Mar'27 Jun'27 Sep'27 Dec'27 Mar'28
```

---

## 6. User Acquisition Channels

| Channel | Cost/User | Volume | Time to Result | Strategy |
|---------|-----------|--------|----------------|----------|
| SimpleX communities | $0 | 500 | 1 month | Post in existing channels, offer beta access |
| Telegram privacy groups | $0.50 | 2,000 | 2 months | Value posts, not spam; target @privacy, @crypto groups |
| DAO forums (Discourse) | $1 | 500 | 3 months | Proposals to DAOs ("Run your treasury on sovereign node") |
| Reddit (r/privacy, r/crypto) | $2 | 3,000 | 3 months | Build-in-public posts, technical deep-dives |
| Product Hunt launch | $0 | 1,000 | Launch day | Prepare 6 months in advance, rally community |
| Twitter/X build in public | $0 | 2,000 | 6 months | Daily dev logs, economy metrics, privacy education |
| Referral program | $5 | 5,000 | 6 months | Free month for referrer + referee |
| Privacy conferences | $50 | 500 | 9 months | Booth + talk: "Digital Nations 101" |
| B2B cold outreach | $100 | 100 | 9 months | Personalized proposals to DAO treasuries |
| Press / PR | $500 | 1,000 | 12 months | "First silver-backed messenger" → mainstream tech press |

### Viral Loop Design
```
User joins → creates address → shares QR → invites friend
→ friend joins → gets bonus ng → invites 2 more
→ each referral = 100 ng bonus (minted from reserve, capped at 10K referrals/user)
```

---

## 7. Technical Roadmap Supporting Growth

### Month 1-3: Foundation
- [ ] PWA web client (REST + SSE, no native build required)
- [ ] Fiat on-ramp (Stripe → USDC → swap to ng)
- [ ] Email/password auth (in addition to mnemonic)
- [ ] Rate limit scaling (ready for 100 concurrent users)
- [ ] Analytics pipeline (product analytics, not just system metrics)

### Month 4-9: Mobile & Scale
- [ ] iOS app (Swift, wraps API)
- [ ] Android app (Kotlin, wraps API)
- [ ] Push notification service
- [ ] On-chain PoR (TON smart contract)
- [ ] TON liquid staking derivative for ng
- [ ] Horizontal scaling: multi-region node federation
- [ ] CDN for radio streaming

### Month 10-18: Network Effects
- [ ] P2P node federation (mesh of sovereign nodes)
- [ ] DAO treasury module (on-chain governance)
- [ ] ParanoidX Pro (desktop app with full VPN)
- [ ] Channel monetization (creators earn from subscribers)
- [ ] RWA marketplace (tokenized real estate, metals, bonds)
- [ ] Cross-chain atomic swaps (BTC, ETH, TON)
- [ ] AI monetary policy governor

---

## 8. Team Scaling Plan

| Role | Phase 1 (Mo 1-3) | Phase 2 (Mo 4-9) | Phase 3 (Mo 10-18) |
|------|------------------|------------------|--------------------|
| Founder/CEO | 1 | 1 | 1 |
| Go backend | 1 | 2 | 5 |
| Mobile (iOS/Android) | 0 | 2 | 4 |
| Frontend (Web) | 0 | 1 | 2 |
| AI/ML | 0 | 1 | 2 |
| Design/UX | 0 | 1 | 2 |
| Marketing/Growth | 0 | 1 | 3 |
| Sales/BizDev | 0 | 0 | 3 |
| Operations/Legal | 0 | 0.5 | 1 |
| Security | 0 | 0.5 | 1 |
| **Total** | **2** | **10** | **24** |

Monthly burn: $5K → $80K → $250K

---

## 9. Risk Mitigation

| Risk | Probability | Impact | Mitigation |
|------|------------|--------|------------|
| Zero users at launch | High | Critical | Pre-sale 100 lifetime memberships, build community pre-launch |
| SimpleX protocol changes | Medium | High | Fork protocol (simplex-fork is already prepared) |
| Mobile app rejection by App Store | Medium | High | PWA fallback, TestFlight parallel distribution |
| Regulatory action on crypto features | Medium | High | Non-custodial by design, geo-distributed nodes |
| Competitor enters space | Medium | Medium | Moats: silver economy, ParanoidX, P2P relay network effects |
| Funding winter / can't raise | Medium | Critical | Bootstrap with subscription revenue, lean team |
| Single developer bus factor | High | Critical | Documentation, hire in Phase 1, pair programming |

---

## 10. Success Metrics Dashboard

### North Star Metric
**Monthly Active Users (MAU) sending at least 1 message**

### Leading Indicators
- Signups per day
- Invites sent per user (viral coefficient K)
- Time to first message (activation)
- 7-day retention
- MRR growth rate
- NPS score

### Lagging Indicators
- Total users
- Monthly Recurring Revenue
- Valuation
- Silver reserve / circulation ratio
- Number of active nodes

### Target Dashboard (Month 18)

```
MAU: 8,000
Signups/day: 55
Viral coefficient K: 1.4
Activation: <5 min to first message
7-day retention: 40%
30-day retention: 25%
MRR: $485K
MRR growth: 15% month-over-month
NPS: 45
Valuation: $50-100M
```

---

## 11. Immediate Action Items (Next 30 Days)

| # | Action | Owner | Cost |
|---|--------|-------|------|
| 1 | Form legal entity (LLC in privacy-friendly jurisdiction) | Founder | $2K |
| 2 | Set up Stripe account + LemonSqueezy for subscriptions | Founder | $0 |
| 3 | Build PWA web client (Flutter web without dart:io) | Dev | $5K |
| 4 | Create landing page: `simplexnode.io` or `stmaria.org/join` | Dev | $1K |
| 5 | Pre-sale: 100 lifetime memberships @ $100 | Founder | $10K raised |
| 6 | Post to 10 SimpleX/Telegram communities | Founder | $0 |
| 7 | Record 60-second demo video | Founder | $0 |
| 8 | Set up analytics (Plausible + PostHog) | Dev | $0 |
| 9 | Onboard 3 beta testers (friends, privacy community) | Founder | $0 |
| 10 | Apply to SimpleX partner program | Founder | $0 |

**Total cost Month 1:** ~$8,000 (or $0 if PWA deferred)

---

## 12. Conclusion

simplex-node has a unique position:
- **Only** complete sovereign node stack (chat + economy + AI + privacy)
- **Only** silver-backed digital messenger
- **Only** Tor-native sovereign network with 100+ API endpoints
- ParanoidX enterprise privacy suite as defensive moat

**The goal is clear:** Convert technical capability into user adoption through:
1. **Product** — Mobile-first experience with fiat on-ramp
2. **Marketing** — Privacy community + DAO outreach + viral referrals
3. **Monetization** — Tiered subscriptions + transaction fees + enterprise
4. **Funding** — Pre-seed → Seed → Series A aligned with milestones

**If we reach 10K users and $500K MRR within 18 months, valuation of $50-100M is realistic** — a 30-100× return on current value.

---

*"Build the network. The network is the nation."*
