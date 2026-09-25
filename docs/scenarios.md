# Scenario examples

This page collects example scenario descriptions that can be supplied to the `sap-iac.scenario` command, which is the step where you describe your application in plain language. Each entry is a self-contained description you can copy, paste, and adapt to your own requirements.

The scenarios are ordered by increasing complexity, from a single application in one subaccount to a governed enterprise landing zone. For a complete, command-by-command run-through that uses one of these descriptions end to end, see the [usage walkthrough](walkthrough.md).

!!! tip "Adapt, don't copy verbatim"
    Each description is a starting point. Replace the application type, services, regions, and environment names with your own before running `sap-iac.scenario`.

## 1. Single application

**Complexity:** Introductory — one application, one subaccount, one environment.

A minimal end-to-end case covering a single application with its core persistence and authentication services.

> "An internal HR leave-request app built with CAP (Node.js) on Cloud Foundry. It stores leave requests in SAP HANA Cloud and authenticates employees via XSUAA. Single environment, one subaccount, region eu10."

## 2. Multiple environments

**Complexity:** Common — one application promoted across three environments under shared guardrails.

Introduces environment isolation and governance: the same application is deployed to separate dev, test, and prod subaccounts, each subject to shared naming and service-plan rules.

> "A customer feedback REST API built with CAP (Node.js), backed by SAP HANA Cloud with XSUAA authentication. Deploy it across three isolated environments — dev, test, and prod — each in its own subaccount, all following shared naming and service-plan guardrails."

## 3. External integrations

**Complexity:** Intermediate — outbound connectivity to on-premise and third-party systems.

Adds the Destination and Connectivity services, OAuth 2.0 client-credentials authentication, and access to an on-premise backend through Cloud Connector.

> "An order-sync service (CAP, Node.js) that reads sales orders from an S/4HANA backend and pushes fulfilment updates to an external shipping provider. Use the Destination service with destinations `s4-orders` and `shipping-api`, both OAuth2ClientCredentials; reach S/4HANA via Cloud Connector (Connectivity service)."

## 4. Multi-application directory

**Complexity:** Advanced — several applications organised under a shared directory hierarchy.

Covers a directory-based account structure with multiple applications that share common services while retaining their own subaccounts and role collections.

> "A retail platform with three apps — `storefront` (CAP, Node.js), `admin` (CAP, Node.js), and `analytics` (Python) — sharing SAP HANA Cloud and XSUAA. Organise them under a directory `acme` with one subaccount per app per environment."

## 5. Enterprise landing zone

**Complexity:** Expert — a governed, multi-business-unit landing zone.

The most comprehensive case: a multi-level directory tree spanning several business units, centralised identity through a custom OIDC provider, organisation-wide role collections, and enforced cost controls.

> "A corporate SAP Business AI Platform landing zone for three business units — retail, finance, and logistics — each running two apps across dev, test, and prod. Central identity via a custom OIDC provider, org-wide role collections, EU-only regions, and mandatory cost-centre tagging."
