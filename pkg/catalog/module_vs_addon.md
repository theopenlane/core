# Modules vs. Add‑ons

| Kind       | What it is                                 | Typical price | Core to the product? | UI placement                 | Examples                 |
|------------|--------------------------------------------|---------------|----------------------|------------------------------|--------------------------|
| **module** | A first‑class, standalone slice of the platform. Customers normally subscribe to **at least one** module to get value. | `$20–$100 / mo` | **Yes** – at least one required | Primary cards in signup & pricing page | `compliance`, `trust_center` |
| **addon**  | An optional enhancement that augments a module. Often usage‑based or small flat fee. | `$1–$10 / mo`  | No – opt‑in            | “Extras / Marketplace” or Billing settings | `vanity_domain`, extra seats |

### Why we keep the lists separate

* **Positioning:** Modules appear in marketing copy as base offerings; add‑ons are upsells.
* **Off‑boarding:** Cancelling the last module should close the subscription; removing an add‑on should not.
* **Visibility controls:** Add‑ons are frequently `beta` or `private` audience.
* **Pricing UI:** Front‑end renders modules and add‑ons in distinct sections for clarity.

Implementation‑wise, the two kinds are identical Go structs; the separation only affects UX.
