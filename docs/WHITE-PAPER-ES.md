# Saint Mary Liberty Island — White Paper (Español)

## La Primera Nación Digital Soberana Privada del Mundo con Economía Respaldada por Plata

**Versión:** 1.0 | **Fecha:** Junio 2026 | **Nombre clave:** "The Isle" / "Остров"
**Contacto:** admin@stmaria.org | **Protocolo:** SimpleX + Tor

---

## Resumen Ejecutivo

Saint Mary Liberty Island está construyendo la infraestructura para una **nación digital soberana** — una red privada y cifrada de comunicación y economía donde las personas pueden comunicarse, realizar transacciones y gobernarse sin vigilancia, censura o riesgo de intermediarios. En su núcleo está **simplex-node**, un único binario Go que impulsa todo el stack: mensajería privada (SimpleX/Tor), una moneda digital respaldada por plata (Liquid Taler), política económica gobernada por IA (Steward), herramientas de pago para comerciantes (POS) y puentes cripto multicadena (TRON/TON).

No somos una startup blockchain. Somos una **plataforma de soberanía digital completa** — el sistema operativo para una nación sin territorio.

---

## 1. El Problema

### 1.1 La Privacidad Está Muerta
Cada mensaje, cada transacción, cada movimiento financiero es rastreado, registrado y monetizado. El cifrado de extremo a extremo está bajo ataque globalmente. Los «cinco ojos», «nueve ojos», «catorce ojos» — el aparato de vigilancia crece diariamente.

### 1.2 El Dinero No Es Sólido
- Las monedas fiduciarias pierden 2-7% de poder adquisitivo anual
- Las CBDC amenazan con control programable del dinero
- Las criptomonedas son volátiles, complejas y rastreables
- La plata física es impráctica para pagos digitales

### 1.3 Las Naciones Digitales No Existen
Ningún proyecto existente combina **mensajería privada + stablecoin respaldada por plata + autocustodia + herramientas comerciales + gobernanza de IA** en una plataforma integrada. Los usuarios deben juntar 5-10 servicios diferentes, cada uno con su propio modelo de seguridad y riesgo de vigilancia.

---

## 2. La Solución: The Isle

### 2.1 Descripción General de la Arquitectura

```
                  ┌─────────────────────────────────────┐
                  │          The Isle (simplex-node)     │
                  │  ┌─────────┐  ┌──────────┐          │
                  │  │ Gateway │  │  Core    │          │
                  │  │ Telegram│  │ Economy  │          │
                  │  │ Signal  │  │ Treasury │          │
                  │  │ WhatsApp│  │ Vault    │          │
                  │  │ Matrix  │  │ POS      │          │
                  │  │ Discord │  │ Steward  │          │
                  │  └─────────┘  └──────────┘          │
                  │       │             │               │
                  │  ┌────▼─────────────▼───────┐       │
                  │  │  Blockchain Bridges       │       │
                  │  │  TRON (USDT) | TON (Jetton)│      │
                  │  │  Bitcoin | Ethereum | Solana │     │
                  │  └────────────────────────────┘       │
                  └───────────────────────────────────────┘
```

### 2.2 Componentes Clave

| Componente | Tecnología | Propósito |
|-----------|-----------|-----------|
| **Mensajería** | SimpleX SMP + XFTP + Tor | Comunicación E2EE privada, transferencia de archivos |
| **Voz/Video** | WebRTC + Coturn TURN | Llamadas cifradas a través de relay oculto |
| **Moneda** | Liquid Taler (ng) | Unidad digital respaldada por plata |
| **Almacenamiento** | Vault (2GB E2EE) | Almacén de archivos autocustodial |
| **Gobernanza** | AI Steward | Política económica constitucional basada en IA |
| **Pagos** | POS Terminal | Facturas comerciales, vales, códigos QR |
| **Puentes Cripto** | TRON/TON → ng | Entrada desde cadenas externas |
| **Asistente IA** | AskSteward (Ollama) | LLM autoalojado (gemma4) |
| **Flota de Bots** | Telegram (Go) | Interfaces interactivas de administración y usuario |

### 2.3 Ventaja Competitiva

Ningún proyecto existente ocupa esta intersección única:

| Necesidad | Soluciones | Nuestra Ventaja |
|-----------|-----------|-----------------|
| Mensajería privada | Signal, SimpleX, Telegram | Sin teléfono/correo electrónico |
| Moneda respaldada por plata | Kinesis, Swiss America | Autosoberana, no custodial |
| Gobernanza de IA | — | Steward de IA constitucional |
| Herramientas comerciales | Square, Stripe | 1% de comisión, prioridad a la privacidad |
| Puente multicadena | Varios DEX | Integrado en plataforma soberana |

**Brecha Competitiva:** Combinamos TODO esto en UN nodo soberano que TÚ controlas.

---

## 3. La Economía de Liquid Taler

### 3.1 Política Monetaria

- **1 Liquid Taler (TLR)** = 31,103,480,000 ng (nanogramos) = 1 onza troy de plata al contado
- **Respaldo en plata:** 70% de plata física en reserva
- **Prima de utilidad:** 30% — valor de la red (velocidad, privacidad, conveniencia)
- **Comisión máxima:** 4.20% (el tesoro toma 2.28%, como máximo)
- **Comisión del tesoro:** Exactamente 228 BPS (2.28%) en cada transacción — sin excepciones

### 3.2 Modelo de Emisión (70 oz → 100 TLR)

Cuando se inicia una ronda de plata:

| Asignación | % | Destinatario |
|-----------|---|-----------|
| Inversor | 70.0% | La parte que pagó USDT |
| Tesoro | 2.4% | Presupuesto operativo del nodo |
| Fondo de Dividendos | 14.1% | Titulares de billetes (por peso de rareza) |
| Compra de Plata | 3.0% | Futuras compras de plata |
| Fondo de Subasta | 4.5% | Liquidez de la casa de subastas |
| Fondo de Recompra | 6.0% | Reserva de recompra de billetes |

### 3.3 Mecanismos Deflacionarios

- **Quema del tesoro:** 20-40% en niveles altos de tesoro
- **Quema de suscripciones:** 30% de ingresos por suscripciones
- **Quema de subastas:** 5% de comisiones de subasta
- **Quema de publicidad:** 20% de compras de etiquetas publicitarias

Estas fuerzas deflacionarias crean una presión de apreciación incorporada en el ng.

---

## 4. ICO: Venta de Tokens Génesis

### 4.1 El Jetton ARGENTUM (TON)

**ARGENTUM** es un Jetton de la blockchain TON vinculado 1:1 con Liquid Taler (ng):

- Símbolo: **ARGENTUM**
- Decimales: 9 (igual que ng)
- Respaldo: 70% plata física
- Comisión de swap: 0.5% (50 BPS)
- Swap mínimo: 1,000,000 ng (~$0.0024)

El contrato maestro Jetton se desplegará en la red principal de TON, permitiendo:
- Comerciar en DEX de TON (STON.fi, DeDust)
- Mantener en cualquier billetera TON (Tonkeeper, Wallet)
- Intercambiar de vuelta a ng en The Isle en cualquier momento (1:1)

### 4.2 Niveles de Venta

| Nivel | Inversión Mínima | Bono | Vesting |
|------|---------------|-------|---------|
| **Genesis Angel** | 100,000 USDT | 30% | 6 meses cliff, 12 meses lineal |
| **Major Investor** | 10,000 USDT | 20% | 3 meses cliff, 9 meses lineal |
| **Minor Investor** | 1,000 USDT | 10% | 1 mes cliff, 6 meses lineal |
| **Citizen** | 100 USDT | 5% | Sin cliff, 3 meses lineal |

### 4.3 Asignación de Fondos

| Uso | % de Recaudación |
|-----|-----------------|
| Compra de Reserva de Plata | 50% |
| Fondos de Liquidez TON Jetton | 15% |
| Fondo de Desarrollo (2 años) | 15% |
| Marketing y Ecosistema | 10% |
| Legal y Cumplimiento | 5% |
| Reserva (operaciones) | 5% |

### 4.4 Distribución de Tokens (Suministro Máximo: 1,000,000 TLR)

| Categoría | % | TLR | Vesting |
|-----------|---|-----|---------|
| Venta ICO | 40% | 400,000 | Según calendario de niveles |
| Reserva del Tesoro | 20% | 200,000 | Controlado por DAO |
| Reserva de Plata | 15% | 150,000 | Correspondiente a plata física |
| Equipo y Asesores | 10% | 100,000 | 2 años lineal |
| Fondo de Ecosistema | 10% | 100,000 | Subvenciones DAO |
| Airdrops Comunitarios | 5% | 50,000 | Después del ICO |

---

## 5. Puentes Cripto Multicadena

### 5.1 Arquitectura

```
 Cadena Externa         Capa de Puente          The Isle
 ┌──────────┐       ┌──────────────┐       ┌──────────┐
 │ TRON     │──────▶│ USDT Monitor  │──────▶│ Treasury │
 │ USDT     │       │ (TronGrid)   │       │ Ledger   │
 └──────────┘       └──────────────┘       └──────────┘
 ┌──────────┐       ┌──────────────┐       ┌──────────┐
 │ TON      │◀─────▶│ ARGENTUM     │◀─────▶│ ng       │
 │ Jetton   │       │ Swap Engine  │       │ Ledger   │
 └──────────┘       └──────────────┘       └──────────┘
 ┌──────────┐       ┌──────────────┐       ┌──────────┐
 │ Bitcoin  │──────▶│ BTC Bridge    │──────▶│ Treasury │
 │ (futuro) │       │ (Atomic Swap)│       │ Ledger   │
 └──────────┘       └──────────────┘       └──────────┘
 ┌──────────┐       ┌──────────────┐       ┌──────────┐
 │ Ethereum │──────▶│ ETH Bridge    │──────▶│ Treasury │
 │ (futuro) │       │ (LayerZero)  │       │ Ledger   │
 └──────────┘       └──────────────┘       └──────────┘
 ┌──────────┐       ┌──────────────┐       ┌──────────┐
 │ Solana   │──────▶│ SOL Bridge    │──────▶│ Treasury │
 │ (futuro) │       │ (Wormhole)   │       │ Ledger   │
 └──────────┘       └──────────────┘       └──────────┘
```

### 5.2 Puentes Actuales (En Vivo)

| Cadena | Activo | Mecanismo | Estado |
|--------|--------|-----------|--------|
| **TRON** | USDT TRC20 | TronGrid polling, registro automático | ✅ En Vivo (Cycle 30) |
| **TON** | ARGENTUM Jetton | Swap esqueleto, contrato pendiente | 🔄 Pre-lanzamiento |

### 5.3 Puentes Planificados (Fase B1 — Post-ICO)

| Cadena | Activo | Mecanismo | Esfuerzo |
|--------|--------|-----------|----------|
| **Bitcoin** | BTC | Atomic swaps + Lightning Network | 4 semanas |
| **Ethereum** | ETH, USDC, USDT | LayerZero OFT / Chainlink CCIP | 6 semanas |
| **Solana** | SOL, USDC | Wormhole bridge | 4 semanas |
| **Polygon** | MATIC, USDC | Polygon PoS bridge | 3 semanas |
| **Arbitrum** | ETH, USDC | Arbitrum bridge | 3 semanas |
| **Base** | ETH, USDC | Base bridge (Coinbase) | 2 semanas |
| **BSC** | BNB, USDT | BSC bridge | 3 semanas |

### 5.4 Junta de Puentes (Gobernanza DAO)

Cada puente será gobernado por una **Junta de Puentes** — un comité multi-firma elegido por los tenedores de ARGENTUM:

- **Validadores de puente:** 5-7 miembros, elegidos trimestralmente
- **Umbral de firma:** 4/7 multi-firma
- **Comisión de puente:** 0.1% por transacción entre cadenas (va al fondo de dividendos)
- **Pausa de emergencia:** 3/7 pueden pausar cualquier puente
- **Ciclo de auditoría:** Verificación PoR mensual por auditores electos

La Junta de Puentes es el primer paso hacia la gobernanza completa de DAO de La Isla.

---

## 6. Bots de Mensajería Multiplataforma

### 6.1 Arquitectura

```
                    Gateway (internal/gateway/)
                    ┌────────────────────────┐
                    │     MultiSender         │
                    │  Sender interface       │
                    └────────┬───────────────┘
                             │
         ┌───────────────────┼───────────────────┐
         ▼                   ▼                   ▼
  ┌────────────┐    ┌────────────┐    ┌────────────┐
  │ Telegram   │    │ WhatsApp   │    │ Signal     │
  │ Sender     │    │ Sender     │    │ Sender     │
  └────────────┘    └────────────┘    └────────────┘
  ┌────────────┐    ┌────────────┐    ┌────────────┐
  │ Matrix     │    │ Discord    │    │ SimpleX    │
  │ Sender     │    │ Sender     │    │ Sender     │
  └────────────┘    └────────────┘    └────────────┘
```

### 6.2 Plan de Implementación (Fase B2 — 12 semanas)

| Plataforma | Biblioteca/API | Método de Autenticación | Esfuerzo |
|-----------|---------------|------------------------|----------|
| **Telegram** | Bot API (existente) | Token de bot | ✅ Hecho |
| **WhatsApp** | WhatsApp Business API | Teléfono + clave API | 2 semanas |
| **Signal** | signal-cli REST API | Número registrado | 2 semanas |
| **Matrix** | Matrix Client-Server API | Token de acceso | 3 semanas |
| **Discord** | Discord Bot API | Token de bot | 2 semanas |
| **SimpleX** | Puente WS CLI (pendiente) | Cola SMP | 3 semanas |

---

## 7. Gobernanza: AI Steward → DAO

### Fase 1: AI Steward (Actual — Cycle 38)
- **Constitución:** 16 reglas fijas con límites mín/máx/objetivo
- **Monitor:** Recopilación de métricas cada 60 segundos (15+ métricas)
- **Analizador:** Detección de desviaciones de 3 niveles (menor/ mayor/crítico)
- **Motor de decisiones:** Ajuste automático (menor), notificar admin (mayor), requerir consenso (crítico)
- **Parámetros dinámicos:** 9 parámetros económicos ajustables persistentes en disco

### Fase 2: IA + Junta Humana (Post-ICO)
- Junta de Puentes electa (5-7 miembros)
- Consejo de auditores (elegido por tenedores de billetes)
- IA hace propuestas, junta humana confirma decisiones críticas
- Votaciones mensuales de gobernanza

### Fase 3: DAO Completo (Objetivo: 2027)
- Voto ponderado por tenencia de ARGENTUM
- Sistema de propuestas en cadena
- DAO de gestión del tesoro
- Enmiendas constitucionales mediante referéndum

---

## 8. Hoja de Ruta

### Fase A (Ahora — Completado)
- ✅ Modelo de respaldo 70% plata
- ✅ Monitor TRON USDT real
- ✅ Núcleo AI Steward (constitución, monitor, analizador)
- ✅ Parámetros económicos dinámicos
- ✅ Terminal POS con facturas QR
- ✅ Protocolo Royal→Sub
- ✅ PDF de billetes con doble firma Ed25519
- ✅ Onboarding y niveles de suscripción
- ✅ Módulo de unificación de puerta de enlace
- ✅ 80+ endpoints API
- ✅ 18/18 paquetes Go con pruebas verdes
- ✅ 4 bots de Telegram

### Fase B1 (Post-ICO — Q3 2026)
- 🔄 Despliegue de ARGENTUM TON Jetton + pools de liquidez
- 🔄 Contratos inteligentes ICO + vesting
- 🔄 Puente Bitcoin atomic swap
- 🔄 Puente Ethereum LayerZero
- 🔄 Puente Solana Wormhole
- 🔄 Elección de la Junta de Puentes
- 🔄 Panel PoR público

### Fase B2 (Q4 2026)
- 🔄 Integración WhatsApp Business API
- 🔄 Puente Signal messenger
- 🔄 Federación Matrix
- 🔄 Bot de Discord para marketplace
- 🔄 App móvil Flutter (iOS + Android)
- 🔄 Soporte nativo WebRTC móvil

### Fase C (2027)
- 🔄 Gobernanza DAO completa
- 🔄 Migración a PostgreSQL (desde JSON)
- 🔄 Escalabilidad a 100,000 usuarios
- 🔄 Red multi-franquicia
- 🔄 Plataforma de tokenización de activos del mundo real
- 🔄 Redención física en plata

---

## 9. Utilidad del Token

| Utilidad | ARGENTUM (TON) | Liquid Taler (ng) |
|---------|---------------|--------------------|
| Reserva de valor respaldada por plata | ✅ 1:1 con ng | ✅ 70% respaldo físico |
| Comisiones de transacción | ❌ | ✅ 2.28% tesoro |
| Votos de gobernanza | ✅ Ponderados | ❌ |
| Valor entre cadenas | ✅ Comercio en DEX TON | ❌ Solo interno |
| Dividendos | ✅ (vía ng) | ✅ (en billetes) |
| Pagos POS | ❌ | ✅ Pagos a comerciantes |
| Staking/DeFi | ✅ (ecosistema TON) | ❌ |

---

## 10. Factores de Riesgo

| Riesgo | Mitigación |
|--------|-----------|
| Volatilidad de la plata | 70% de respaldo amortigua caída del 30% |
| Caída de la red TON | Estrategia multicadena, no depende de TON |
| Fallo de gobernanza de IA | Anulación manual + límites constitucionales |
| Incertidumbre regulatoria | Descentralizado, sin KYC |
| Desanclaje de USDT | Diversificación a USDC, DAI, cripto nativas |
| Cambios en API de SimpleX | Versión CLI fijada en Docker |
| Fallo de disco | Copias de seguridad diarias a USB + nube |

---

## 11. Equipo

The Isle es construido por un equipo distribuido de ingenieros de privacidad, economistas monetarios e investigadores de IA que operan bajo el proyecto soberano **Saint Mary Liberty Island**. Roles clave:

- **King Tomas** — Fundador, arquitecto de sistemas
- **AI Steward** — Gobernanza constitucional de IA
- **AskSteward (AI)** — Asistente público de IA
- **Inquisitor (AI)** — Automatización de QA de ciclos de producción

El desarrollo se financia a través de ingresos por suscripciones, comisiones del tesoro y el Genesis ICO.

---

## 12. Estructura Legal

Saint Mary Liberty Island opera como una **entidad digital soberana** bajo el paraguas de stmaria.org / markbank.org. El proyecto no es una empresa, no es una fundación — es una **red de nodos soberanos** gobernada por software.

- **Sin KYC/AML**: El protocolo es autosoberano
- **Sin riesgo de custodia**: Los usuarios controlan sus claves
- **No es un valor**: Liquid Taler es un token de utilidad respaldado por plata física
- **Opinión legal**: Se ha contratado asesoría para la estructuración del ICO (Gibraltar/Dubái/Suiza)

---

## 13. Llamada a la Acción

### Para Inversores Minoritarios ($100+)
Participe en el nivel Citizen del Genesis ICO. Reciba Jettons ARGENTUM en TON, con vesting de 3 meses y acceso completo al ecosistema de The Isle.

- **Sitio web:** https://stmaria.org
- **Nodo:** simplex-node en localhost:8080 (accesible por onion)
- **Telegram:** @AskSteward_bot | @torquemada878_bot

### Para Inversores Mayoritarios ($10,000+)
Lista blanca para el nivel Major Investor. Comunicación directa con King Tomas. Elegibilidad para la Junta de Puentes. Primer acceso a licencias de nodos franquiciados.

### Para Inversores Institucionales ($100,000+)
Nivel Genesis Angel. Calendario de vesting personalizado. Nodo validador de puente dedicado. Asignación preferente en futuras rondas de franquicias.

---

*"Mensajes privados. Dinero privado. Respaldo en plata."*

**Saint Mary Liberty Island**
stmaria.org — markbank.org
Junio 2026
