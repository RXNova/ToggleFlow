# ToggleFlow Infrastructure

## Domain Registration
**Provider:** one.com
**Domain:** toggleflow.io
**Nameservers:** Delegated to Cloudflare (see DNS Management below)

---

## DNS Management
**Provider:** Cloudflare (free plan)
**Records:**
| Type | Name | Target | Proxy |
|---|---|---|---|
| CNAME | `@` | `1b17d2d5-9fcc-4af8-abe4-b1a1b7248c7e.cname.koyeb.app` | DNS only |
| CNAME | `www` | `1b17d2d5-9fcc-4af8-abe4-b1a1b7248c7e.cname.koyeb.app` | DNS only |

---

## Application Hosting
**Provider:** Koyeb (free plan)
**Region:** Washington D.C. (was)
**App:** toggleflow
**Service:** toggleflow
**URL:** https://toggleflow.io
**Fallback URL:** https://toggleflow-toggleflow-fcd1c64a.koyeb.app
**Instance:** Free (0.1 vCPU, 512 MB RAM, 2 GB SSD)
**Source:** GitHub — `toggle-flow/ToggleFlow`, branch `fly-demo`
**Scales to zero after:** 1 hour of inactivity (mitigated by UptimeRobot)

---

## Uptime Monitoring
**Provider:** UptimeRobot (free plan)
**Monitor:** https://toggleflow.io/health
**Interval:** every 5 minutes
**Purpose:** Keeps Koyeb instance warm — prevents cold starts

---

## Health Check
**Endpoint:** `GET https://toggleflow.io/health`
**Response:** `{"status":"ok"}`
**Also monitored by:** Koyeb internal health check (same endpoint, every 30s)
