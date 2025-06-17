# Product Catalog

each of our "modules" is a stripe product

each stripe product has at least 2 prices usually, the difference between the interval (monthly or annually)

an organization has 1 customer record

a customer will only have 1 subscription with many products and prices

you cannot co-mingle nonthly and annual pricing


===========================

Main differences between "tier" and "module" ->

In the previous setup the subscription was either "active" or "not active" and it would only have 1 tier associated

out of the total list of features, each tier included the list of features from the lower tier, and then it's tier specific features with the highest tier holding the total inventory of features

In the new model, a customer will have 1 subscription but they could have many products some in an active state or not (previously the "tier" was a product; our modules are also "products" so its a more granular representation)
