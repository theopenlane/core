
| Resource| Definition |
-------------------------
Customer	Represents a customer who purchases a subscription. Use the Customer object associated with a subscription to make and track recurring charges and to manage the products that they subscribe to.
Entitlement	Represents a customer’s access to a feature included in a service product that they subscribe to. When you create a subscription for a customer’s recurring purchase of a product, an active entitlement is automatically created for each feature associated with that product. When a customer accesses your services, use their active entitlements to enable the features included in their subscription.
Feature	Represents a function or ability that your customers can access when they subscribe to a service product. You can include features in a product by creating ProductFeatures.
Invoice	A statement of amounts a customer owes that tracks payment statuses from draft through paid or otherwise finalized. Subscriptions automatically generate invoices.
PaymentIntent	A way to build dynamic payment flows. A PaymentIntent tracks the lifecycle of a customer checkout flow and triggers additional authentication steps when required by regulatory mandates, custom Radar fraud rules, or redirect-based payment methods. Invoices automatically create PaymentIntents.
PaymentMethod	A customer’s payment instruments that they use to pay for your products. For example, you can store a credit card on a Customer object and use it to make recurring payments for that customer. Typically used with the Payment Intents or Setup Intents APIs.
Price	Defines the unit price, currency, and billing cycle for a product.
Product	A good or service that your business sells. A service product can include one or more features.
ProductFeature	Represents a single feature’s inclusion in a single product. Each product is associated with a ProductFeature for each feature that it includes, and each feature is associated with a ProductFeature for each product that includes it.
Subscription	Represents a customer’s scheduled recurring purchase of a product. Use a subscription to collect payments and provide repeated delivery of or continuous access to a product.

