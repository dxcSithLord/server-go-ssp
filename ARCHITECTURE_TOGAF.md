# SQRL SSP - TOGAF Architecture Documentation
**Framework:** The Open Group Architecture Framework (TOGAF)
**Version:** 9.2
**Project:** SQRL Server-Side Protocol Implementation
**Date:** November 18, 2025

---

## Executive Summary

This document provides a TOGAF-aligned architecture view of the SQRL SSP (Secure QR Login Server-Side Protocol) implementation, mapping business objectives to functional requirements and technical architecture.

---

## TOGAF Architecture Development Method (ADM) Phases

### Phase A: Architecture Vision

#### Business Objectives

```mermaid
graph TB
    A[Eliminate Password<br/>Vulnerabilities] --> B[Secure Passwordless<br/>Authentication]
    C[Enable Mobile-First<br/>Authentication] --> B
    D[Support Horizontal<br/>Scaling] --> B
    E[Maintain User<br/>Privacy] --> B

    B --> F[SQRL SSP<br/>Implementation]

    F --> G[Reduce Breach Risk]
    F --> H[Improve User Experience]
    F --> I[Lower Support Costs]
    F --> J[Enable Integration]

    style F fill:#e1f5ff
    style B fill:#ffe1e1
    style G fill:#e1ffe1
    style H fill:#e1ffe1
    style I fill:#e1ffe1
    style J fill:#e1ffe1
```

#### Key Stakeholders

| Stakeholder | Concern | Requirement |
|------------|---------|-------------|
| **End Users** | Easy authentication without passwords | QR code scanning, mobile-first |
| **Security Team** | Strong authentication without credential storage | ED25519 cryptography, no password storage |
| **Operations Team** | Scalable, reliable deployment | Horizontal scaling, monitoring |
| **Developers** | Easy integration with existing systems | Pluggable interfaces, clear API |
| **Compliance** | GDPR, security standards | Privacy-preserving, audit trails |

---

## TOGAF Business Architecture

### Business Capability Map

```mermaid
graph LR
    subgraph "Authentication Services"
        A1[Identity Verification]
        A2[Session Management]
        A3[Access Control]
    end

    subgraph "User Management"
        U1[Identity Creation]
        U2[Identity Lifecycle]
        U3[Account Recovery]
    end

    subgraph "Security Operations"
        S1[Cryptographic Operations]
        S2[Threat Detection]
        S3[Audit Logging]
    end

    subgraph "Integration Services"
        I1[API Gateway]
        I2[OAuth2/OIDC Bridge]
        I3[External Auth Providers]
    end

    A1 --> S1
    A2 --> U2
    U1 --> S1
    I1 --> A1
    I2 --> A2

    style A1 fill:#ffe1e1
    style S1 fill:#e1e1ff
    style U1 fill:#e1ffe1
```

### Business Process: User Authentication

```mermaid
sequenceDiagram
    autonumber
    actor User
    participant Browser
    participant SQRL_SSP as SQRL SSP Server
    participant SQRL_Client as SQRL Client
    participant Authenticator
    participant Application

    User->>Browser: Navigate to login page
    Browser->>SQRL_SSP: GET /nut.sqrl
    SQRL_SSP-->>Browser: nut, pag, exp
    Browser->>SQRL_SSP: GET /png.sqrl?nut=X
    SQRL_SSP-->>Browser: QR Code PNG
    Browser->>SQRL_SSP: Poll /pag.sqrl?nut=X&pag=Y
    SQRL_SSP-->>Browser: (pending)

    User->>SQRL_Client: Scan QR Code
    SQRL_Client->>SQRL_SSP: POST /cli.sqrl (signed)
    SQRL_SSP->>SQRL_SSP: Verify ED25519 Signature
    SQRL_SSP->>Authenticator: AuthenticateIdentity()
    Authenticator-->>SQRL_SSP: Redirect URL
    SQRL_SSP-->>SQRL_Client: Success Response

    Browser->>SQRL_SSP: Poll /pag.sqrl?nut=X&pag=Y
    SQRL_SSP-->>Browser: Redirect URL
    Browser->>Application: Redirect to Application
    Application-->>User: Authenticated Session
```

---

## TOGAF Application Architecture

### Application Component Model

```mermaid
graph TB
    subgraph "Presentation Layer"
        P1[Demo Web UI]
        P2[QR Code Generator]
        P3[API Endpoints]
    end

    subgraph "Business Logic Layer"
        B1[Authentication Handler]
        B2[Identity Manager]
        B3[Request Validator]
        B4[Response Builder]
        B5[Signature Verifier]
    end

    subgraph "Integration Layer"
        I1[Authenticator Interface]
        I2[Hoard Interface]
        I3[AuthStore Interface]
        I4[Tree Interface]
    end

    subgraph "Storage Layer"
        S1[Nut Storage<br/>MapHoard/EtcdHoard]
        S2[Identity Storage<br/>MapAuthStore/EtcdAuthStore]
        S3[Nut Generator<br/>RandomTree/GrcTree]
    end

    subgraph "External Integrations"
        E1[Application Auth System]
        E2[User Directory]
        E3[OAuth2/OIDC Provider]
        E4[Git Hosting]
    end

    P1 --> B1
    P2 --> B1
    P3 --> B1

    B1 --> B3
    B1 --> B5
    B1 --> B4
    B1 --> B2

    B2 --> I1
    B1 --> I2
    B2 --> I3
    B1 --> I4

    I1 --> E1
    I2 --> S1
    I3 --> S2
    I4 --> S3

    I1 -.-> E3
    E3 -.-> E4

    style B1 fill:#ffe1e1
    style B5 fill:#e1e1ff
    style I1 fill:#e1ffe1
    style I2 fill:#e1ffe1
    style I3 fill:#e1ffe1
    style I4 fill:#e1ffe1
```

### Application Interaction Model

```mermaid
graph LR
    subgraph "SQRL SSP Core"
        API[SqrlSspAPI]
        CLI[CLI Handler]
        NUT[Nut Handler]
        PNG[PNG Handler]
        PAG[Pag Handler]
    end

    subgraph "Interfaces"
        AUTH[Authenticator]
        HOARD[Hoard]
        STORE[AuthStore]
        TREE[Tree]
    end

    subgraph "Implementations"
        MAP_H[MapHoard]
        ETCD_H[EtcdHoard]
        MAP_A[MapAuthStore]
        ETCD_A[EtcdAuthStore]
        RAND_T[RandomTree]
        GRC_T[GrcTree]
        CUSTOM[Custom Authenticator]
    end

    API --> CLI
    API --> NUT
    API --> PNG
    API --> PAG

    CLI --> AUTH
    CLI --> HOARD
    CLI --> STORE

    NUT --> TREE
    NUT --> HOARD

    HOARD -.->|implements| MAP_H
    HOARD -.->|implements| ETCD_H
    STORE -.->|implements| MAP_A
    STORE -.->|implements| ETCD_A
    TREE -.->|implements| RAND_T
    TREE -.->|implements| GRC_T
    AUTH -.->|implements| CUSTOM

    style API fill:#ffe1e1
    style HOARD fill:#e1ffe1
    style STORE fill:#e1ffe1
    style AUTH fill:#e1ffe1
    style TREE fill:#e1ffe1
```

---

## TOGAF Data Architecture

### Conceptual Data Model

```mermaid
erDiagram
    SQRL_IDENTITY ||--o{ NUT_CACHE : "associated with"
    SQRL_IDENTITY {
        string idk PK "Identity Public Key (ED25519)"
        string suk "Server Unlock Key"
        string vuk "Verify Unlock Key"
        string pidk "Previous Identity Key"
        bool sqrlOnly "SQRL-only authentication"
        bool hardlock "Requires full unlock"
        bool disabled "Account disabled"
        string rekeyed "New identity if rekeyed"
        int btn "Ask button response"
    }

    NUT_CACHE {
        string nut PK "Cryptographic nonce"
        string state "issued/associated/authenticated"
        string remoteIP "Client IP address"
        string originalNut "Original nut"
        string pagNut "Polling nut"
        time expiration "Expiration time"
        bytes lastResponse "Previous response"
    }

    AUTHENTICATION_EVENT {
        string nut FK "Associated nut"
        string idk FK "Identity key"
        time timestamp "Event time"
        string command "SQRL command"
        string remoteIP "Client IP"
        int tif "Transaction flags"
        bool success "Authentication result"
    }

    NUT_CACHE }|--|| AUTHENTICATION_EVENT : "logs"
    SQRL_IDENTITY }|--o{ AUTHENTICATION_EVENT : "performs"
```

### Data Flow Diagram

```mermaid
graph LR
    subgraph "Data Sources"
        CLIENT[SQRL Client]
        BROWSER[Web Browser]
    end

    subgraph "SQRL SSP Processing"
        RECEIVE[Receive Request]
        DECODE[Decode Base64]
        VALIDATE[Validate Signature]
        PROCESS[Process Command]
        ENCODE[Encode Response]
    end

    subgraph "Data Stores"
        HOARD[(Nut Hoard<br/>Ephemeral)]
        AUTH_STORE[(Identity Store<br/>Persistent)]
        LOGS[(Audit Logs<br/>Append-Only)]
    end

    subgraph "Data Outputs"
        RESPONSE[Server Response]
        REDIRECT[Redirect URL]
        AUDIT[Audit Event]
    end

    CLIENT -->|Signed Client Data| RECEIVE
    BROWSER -->|Nut Request| RECEIVE

    RECEIVE --> DECODE
    DECODE --> VALIDATE
    VALIDATE -->|Valid| PROCESS
    VALIDATE -->|Invalid| ENCODE

    PROCESS <-->|Read/Write| HOARD
    PROCESS <-->|Read/Write| AUTH_STORE
    PROCESS -->|Log| LOGS

    PROCESS --> ENCODE
    ENCODE --> RESPONSE
    PROCESS --> REDIRECT
    PROCESS --> AUDIT

    RESPONSE --> CLIENT
    REDIRECT --> BROWSER
    AUDIT --> LOGS

    style VALIDATE fill:#e1e1ff
    style HOARD fill:#ffe1e1
    style AUTH_STORE fill:#ffe1e1
    style LOGS fill:#e1ffe1
```

---

## TOGAF Technology Architecture

### Technology Component Model

```mermaid
graph TB
    subgraph "Runtime Environment"
        GO[Go 1.25.4 Runtime]
        OS[Linux/macOS/Windows]
    end

    subgraph "Application Layer"
        SSP[SQRL SSP Binary]
        CRYPTO[crypto/ed25519<br/>crypto/aes]
        HTTP[net/http Server]
        QR[yeqown/go-qrcode]
    end

    subgraph "Storage Layer"
        MEM[In-Memory Storage]
        ETCD[etcd v3.6.6]
        DB[(PostgreSQL/MySQL<br/>via GORM)]
        REDIS[(Redis<br/>for Hoard)]
    end

    subgraph "Integration Layer"
        OAUTH[OAuth2/OIDC Client]
        AUTHENTIK[Authentik API]
        GOGS[Gogs/Gitea API]
    end

    subgraph "Infrastructure"
        LB[Load Balancer<br/>nginx/HAProxy]
        MONITOR[Prometheus/Grafana]
        LOG[Logging<br/>syslog/CloudWatch]
    end

    OS --> GO
    GO --> SSP
    SSP --> CRYPTO
    SSP --> HTTP
    SSP --> QR

    SSP --> MEM
    SSP -.->|Optional| ETCD
    SSP -.->|Optional| DB
    SSP -.->|Optional| REDIS

    SSP -.->|Optional| OAUTH
    OAUTH -.-> AUTHENTIK
    AUTHENTIK -.-> GOGS

    HTTP --> LB
    SSP --> MONITOR
    SSP --> LOG

    style SSP fill:#ffe1e1
    style CRYPTO fill:#e1e1ff
    style ETCD fill:#e1ffe1
```

### Deployment Architecture (Current State)

```mermaid
graph TB
    subgraph "Client Tier"
        WEB[Web Browser]
        SQRL[SQRL Mobile Client]
    end

    subgraph "Application Tier"
        SERVER[SQRL SSP Server<br/>:8000]
        TREE[RandomTree<br/>Nut Generator]
        HOARD_MAP[MapHoard<br/>In-Memory]
        AUTH_MAP[MapAuthStore<br/>In-Memory]
    end

    WEB -->|HTTPS| SERVER
    SQRL -->|HTTPS| SERVER

    SERVER --> TREE
    SERVER --> HOARD_MAP
    SERVER --> AUTH_MAP

    style SERVER fill:#ffe1e1
    style HOARD_MAP fill:#ffcccc
    style AUTH_MAP fill:#ffcccc

    classDef warning fill:#ffcccc,stroke:#ff0000
    class HOARD_MAP,AUTH_MAP warning
```

*Note: Red indicates NOT suitable for production (in-memory only)*

### Deployment Architecture (Target State with etcd)

```mermaid
graph TB
    subgraph "Client Tier"
        WEB[Web Browser]
        SQRL[SQRL Mobile Client]
    end

    subgraph "Load Balancer Tier"
        LB[nginx/HAProxy<br/>TLS Termination]
    end

    subgraph "Application Tier - Zone A"
        SSP1[SQRL SSP #1<br/>:8000]
    end

    subgraph "Application Tier - Zone B"
        SSP2[SQRL SSP #2<br/>:8000]
    end

    subgraph "Application Tier - Zone C"
        SSP3[SQRL SSP #3<br/>:8000]
    end

    subgraph "Distributed Storage Tier"
        ETCD1[etcd #1<br/>:2379]
        ETCD2[etcd #2<br/>:2379]
        ETCD3[etcd #3<br/>:2379]
    end

    subgraph "Monitoring Tier"
        PROM[Prometheus]
        GRAF[Grafana]
        ALERT[AlertManager]
    end

    WEB --> LB
    SQRL --> LB

    LB --> SSP1
    LB --> SSP2
    LB --> SSP3

    SSP1 --> ETCD1
    SSP1 --> ETCD2
    SSP1 --> ETCD3

    SSP2 --> ETCD1
    SSP2 --> ETCD2
    SSP2 --> ETCD3

    SSP3 --> ETCD1
    SSP3 --> ETCD2
    SSP3 --> ETCD3

    ETCD1 <--> ETCD2
    ETCD2 <--> ETCD3
    ETCD3 <--> ETCD1

    SSP1 --> PROM
    SSP2 --> PROM
    SSP3 --> PROM

    ETCD1 --> PROM
    ETCD2 --> PROM
    ETCD3 --> PROM

    PROM --> GRAF
    PROM --> ALERT

    style LB fill:#e1ffe1
    style SSP1 fill:#ffe1e1
    style SSP2 fill:#ffe1e1
    style SSP3 fill:#ffe1e1
    style ETCD1 fill:#e1e1ff
    style ETCD2 fill:#e1e1ff
    style ETCD3 fill:#e1e1ff
```

---

## Objectives to Requirements Mapping

### Strategic Alignment

```mermaid
graph TD
    subgraph "Strategic Objectives"
        O1[Eliminate Password<br/>Security Risks]
        O2[Enable Passwordless<br/>Future]
        O3[Improve User<br/>Experience]
        O4[Support Business<br/>Growth]
    end

    subgraph "Functional Requirements"
        F1[Cryptographic<br/>Authentication]
        F2[Mobile-First<br/>QR Codes]
        F3[Identity<br/>Management]
        F4[Horizontal<br/>Scaling]
        F5[API<br/>Integration]
    end

    subgraph "Non-Functional Requirements"
        N1[99.9% Uptime]
        N2[Sub-Second<br/>Response Time]
        N3[Security:<br/>ED25519]
        N4[Privacy:<br/>No PII Storage]
        N5[Compliance:<br/>GDPR]
    end

    subgraph "Implementation"
        I1[SQRL Protocol]
        I2[ED25519 Signatures]
        I3[QR Code Generation]
        I4[etcd Distributed Storage]
        I5[Pluggable Interfaces]
        I6[Secure Memory Clearing]
        I7[Safe Logging]
    end

    O1 --> F1
    O1 --> F3
    O2 --> F1
    O2 --> F5
    O3 --> F2
    O3 --> N2
    O4 --> F4
    O4 --> N1

    F1 --> I1
    F1 --> I2
    F2 --> I3
    F3 --> I5
    F4 --> I4
    F4 --> I5
    F5 --> I5

    N3 --> I2
    N3 --> I6
    N4 --> I7
    N4 --> I6
    N5 --> I7

    style O1 fill:#ffe1e1
    style F1 fill:#e1e1ff
    style I2 fill:#e1ffe1
```

### Requirements Traceability Matrix

| Strategic Objective | Functional Requirement | Non-Functional Requirement | Implementation Component | Status |
|---------------------|------------------------|---------------------------|-------------------------|--------|
| **Eliminate Password Risks** | Cryptographic Authentication | Security: ED25519 | cli_handler.go, cli_request.go | ✅ Complete |
| **Eliminate Password Risks** | Identity Management | Privacy: No PII Storage | api.go, SqrlIdentity | ✅ Complete |
| **Passwordless Future** | API Integration | Extensibility: Interfaces | Authenticator, AuthStore, Hoard | ✅ Complete |
| **Passwordless Future** | OAuth2/OIDC Bridge | Standards Compliance | Stage 4: Authentik Integration | 📋 Planned |
| **Improve User Experience** | Mobile-First QR Codes | Performance: <1s Response | handers.go, QR generation | ✅ Complete |
| **Improve User Experience** | One-Step QR Generation | Usability Enhancement | /png.sqrl extension | ✅ Complete |
| **Support Business Growth** | Horizontal Scaling | Availability: 99.9% Uptime | Stage 3: etcd Integration | 📋 Planned |
| **Support Business Growth** | Multi-Datacenter Deployment | Reliability: Auto-failover | etcd cluster (3-5 nodes) | 📋 Planned |
| **Maintain Security** | Secure Memory Clearing | Security: CWE-226 | secure_clear.go | ✅ Complete |
| **Maintain Security** | Safe Logging | Security: CWE-200, CWE-532 | secure_log.go | ✅ Complete |
| **Compliance** | Audit Logging | Compliance: GDPR | Future: audit_log.go | ❌ Not Started |
| **Compliance** | Data Retention Policies | Compliance: Data Minimization | Future: retention policies | ❌ Not Started |

**Legend:**
- ✅ Complete
- 📋 Planned (in upgrade plan)
- ❌ Not Started
- ⚠️ In Progress

---

## Integration Architecture

### Integration Points

```mermaid
graph TB
    subgraph "SQRL SSP Core"
        SSP[SQRL SSP API]
    end

    subgraph "Integration Interfaces"
        AUTH_IF[Authenticator Interface]
        STORE_IF[AuthStore Interface]
        HOARD_IF[Hoard Interface]
    end

    subgraph "Storage Integrations"
        ETCD[etcd v3.6<br/>Distributed KV]
        REDIS[Redis<br/>Session Cache]
        POSTGRES[PostgreSQL<br/>User DB]
        MYSQL[MySQL<br/>User DB]
    end

    subgraph "Authentication Integrations"
        CUSTOM_AUTH[Custom Auth System]
        LDAP[LDAP/AD]
        SAML[SAML IdP]
        OAUTH[OAuth2 Provider]
    end

    subgraph "Identity Providers"
        AUTHENTIK[Authentik<br/>OAuth2/OIDC/SAML]
        KEYCLOAK[Keycloak<br/>Identity Management]
    end

    subgraph "Application Integrations"
        WEB_APP[Web Applications]
        GIT[Gogs/Gitea<br/>Git Hosting]
        API_GW[API Gateway]
    end

    SSP --> AUTH_IF
    SSP --> STORE_IF
    SSP --> HOARD_IF

    HOARD_IF --> ETCD
    HOARD_IF --> REDIS

    STORE_IF --> POSTGRES
    STORE_IF --> MYSQL
    STORE_IF --> ETCD

    AUTH_IF --> CUSTOM_AUTH
    AUTH_IF -.->|via Authentik| AUTHENTIK
    AUTH_IF -.->|via Keycloak| KEYCLOAK

    AUTHENTIK --> LDAP
    AUTHENTIK --> SAML
    AUTHENTIK --> OAUTH

    OAUTH --> WEB_APP
    OAUTH --> GIT
    OAUTH --> API_GW

    style SSP fill:#ffe1e1
    style AUTH_IF fill:#e1ffe1
    style STORE_IF fill:#e1ffe1
    style HOARD_IF fill:#e1ffe1
    style AUTHENTIK fill:#e1e1ff
```

---

## Security Architecture

### Security Zones

```mermaid
graph TB
    subgraph "Public Zone (Internet)"
        CLIENT[SQRL Clients]
        BROWSER[Web Browsers]
    end

    subgraph "DMZ (Demilitarized Zone)"
        LB[Load Balancer<br/>TLS Termination]
        WAF[Web Application<br/>Firewall]
    end

    subgraph "Application Zone (Private Network)"
        SSP1[SQRL SSP Server 1]
        SSP2[SQRL SSP Server 2]
        SSP3[SQRL SSP Server 3]
    end

    subgraph "Data Zone (Highly Restricted)"
        ETCD_CLUSTER[etcd Cluster<br/>Encrypted Storage]
        DB[Database<br/>Encrypted at Rest]
    end

    subgraph "Management Zone"
        MONITOR[Monitoring]
        LOGS[Log Aggregation]
        ADMIN[Admin Console]
    end

    CLIENT -->|HTTPS| LB
    BROWSER -->|HTTPS| LB

    LB --> WAF
    WAF -->|mTLS| SSP1
    WAF -->|mTLS| SSP2
    WAF -->|mTLS| SSP3

    SSP1 -->|Encrypted| ETCD_CLUSTER
    SSP2 -->|Encrypted| ETCD_CLUSTER
    SSP3 -->|Encrypted| ETCD_CLUSTER

    SSP1 -->|Encrypted| DB
    SSP2 -->|Encrypted| DB
    SSP3 -->|Encrypted| DB

    SSP1 --> LOGS
    SSP2 --> LOGS
    SSP3 --> LOGS

    SSP1 --> MONITOR
    SSP2 --> MONITOR
    SSP3 --> MONITOR

    ADMIN -.->|Bastion Host| SSP1

    style LB fill:#ffe1e1
    style WAF fill:#e1e1ff
    style ETCD_CLUSTER fill:#e1ffe1
    style DB fill:#e1ffe1
```

### Security Controls

```mermaid
graph LR
    subgraph "Preventive Controls"
        P1[ED25519 Signatures]
        P2[AES Encryption]
        P3[Input Validation]
        P4[Memory Clearing]
        P5[Safe Logging]
    end

    subgraph "Detective Controls"
        D1[Signature Verification]
        D2[Audit Logging]
        D3[Anomaly Detection]
        D4[Security Scanning]
    end

    subgraph "Corrective Controls"
        C1[Rate Limiting]
        C2[Account Disable]
        C3[Incident Response]
        C4[Patch Management]
    end

    P1 --> D1
    P3 --> D3
    P4 --> D4
    P5 --> D2

    D1 -->|Failure| C2
    D3 -->|Anomaly| C1
    D4 -->|Vulnerability| C4

    style P1 fill:#e1ffe1
    style P4 fill:#e1ffe1
    style D1 fill:#e1e1ff
    style C2 fill:#ffe1e1
```

---

## Migration Architecture (Upgrade Path)

### Staged Migration

```mermaid
graph LR
    subgraph "Stage 0: Current"
        S0[Go 1.25.4<br/>skip2/go-qrcode<br/>MapHoard/MapAuthStore]
    end

    subgraph "Stage 1: QR Library"
        S1[Replace QR Library<br/>yeqown/go-qrcode v2.3.1<br/>Security Improvement]
    end

    subgraph "Stage 2: Test Coverage"
        S2[Increase Tests<br/>29.4% → 80%<br/>Production Readiness]
    end

    subgraph "Stage 3: Distributed Storage"
        S3[Add etcd v3.6.6<br/>EtcdHoard/EtcdAuthStore<br/>Horizontal Scaling]
    end

    subgraph "Stage 4: OAuth2 Bridge"
        S4[Integrate Authentik<br/>OAuth2/OIDC Support<br/>Enterprise Integration]
    end

    subgraph "Stage 5: Git Hosting"
        S5[Add Gitea<br/>Git + SQRL Auth<br/>Developer Platform]
    end

    S0 -->|1 week| S1
    S1 -->|2-3 weeks| S2
    S2 -->|2-3 weeks| S3
    S3 -->|3-4 weeks<br/>OPTIONAL| S4
    S4 -->|2-3 weeks<br/>OPTIONAL| S5

    S1 -.->|Can Deploy| PROD1[Production<br/>Single Server]
    S2 -.->|Can Deploy| PROD2[Production<br/>High Quality]
    S3 -.->|Can Deploy| PROD3[Production<br/>Multi-Server]
    S4 -.->|Can Deploy| PROD4[Production<br/>Enterprise]
    S5 -.->|Can Deploy| PROD5[Production<br/>Full Platform]

    style S1 fill:#e1ffe1
    style S2 fill:#e1ffe1
    style S3 fill:#e1ffe1
    style PROD1 fill:#ffe1e1
    style PROD2 fill:#ffe1e1
    style PROD3 fill:#ffe1e1
```

---

## Performance Architecture

### Performance Requirements

```mermaid
graph TB
    subgraph "Performance Targets"
        P1[Response Time<br/>/nut.sqrl: <100ms]
        P2[Response Time<br/>/cli.sqrl: <200ms]
        P3[Throughput<br/>≥500 QPS]
        P4[Concurrency<br/>≥1000 concurrent users]
    end

    subgraph "Optimization Strategies"
        O1[RandomTree<br/>Pre-generated Nuts]
        O2[Connection Pooling<br/>etcd/Database]
        O3[In-Memory Cache<br/>Hot Identities]
        O4[Horizontal Scaling<br/>Load Distribution]
    end

    subgraph "Monitoring Metrics"
        M1[P50/P95/P99<br/>Latency]
        M2[Requests/Second<br/>Throughput]
        M3[Error Rate<br/>% Failed]
        M4[Saturation<br/>CPU/Memory]
    end

    P1 --> O1
    P2 --> O2
    P3 --> O4
    P4 --> O4

    O1 --> M1
    O2 --> M1
    O3 --> M2
    O4 --> M2

    M1 --> ALERT{Alert if<br/>P95 > 500ms}
    M2 --> ALERT2{Alert if<br/>< 400 QPS}
    M3 --> ALERT3{Alert if<br/>> 1%}

    style P1 fill:#e1ffe1
    style P2 fill:#e1ffe1
    style O1 fill:#e1e1ff
    style O4 fill:#e1e1ff
```

---

## Conclusion

This TOGAF-aligned architecture documentation provides a comprehensive view of the SQRL SSP implementation, mapping business objectives through functional requirements to technical implementation. The architecture supports:

1. **Scalability:** Horizontal scaling via etcd (Stage 3)
2. **Security:** ED25519 cryptography, secure memory, safe logging
3. **Extensibility:** Pluggable interfaces for integration
4. **Maintainability:** Clear separation of concerns, well-tested
5. **Compliance:** Privacy-preserving, audit-capable

The staged upgrade plan (DEPENDENCY_UPGRADE_PLAN.md) provides a safe path to enhance the architecture incrementally, allowing production deployment at multiple checkpoints.

---

**Document Version:** 1.0
**TOGAF Version:** 9.2
**Last Updated:** November 18, 2025
**Status:** Published
