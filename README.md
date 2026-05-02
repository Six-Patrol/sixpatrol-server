# 🛡️ SixPatrol
**The 6-Layer Digital Asset Protection System for Live Broadcasts.**

[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)
[![C++17](https://img.shields.io/badge/C++-17-blue.svg)](https://isocpp.org/)
[![Go](https://img.shields.io/badge/Go-1.21-00ADD8.svg)](https://go.dev/)
[![AI Model: DINOv2](https://img.shields.io/badge/AI-DINOv2-orange.svg)](https://github.com/facebookresearch/dinov2)
[![Google Gemini](https://img.shields.io/badge/Powered_by-Gemini_AI-4285F4.svg)](https://deepmind.google/technologies/gemini/)

> **SixPatrol** is a defense-in-depth anti-piracy platform designed specifically for live sports broadcasters. It solves the critical industry challenge of protecting proprietary digital media from unauthorized redistribution without compromising the zero-latency requirements of a live 4K broadcast.

---

## 📖 Table of Contents
- [The Problem](#-the-problem)
- [The 6-Layer Defense](#-the-6-layer-defense-architecture)
- [Core Features](#-core-features)
- [Tech Stack](#-tech-stack)
- [System Architecture](#-system-architecture)
- [Getting Started](#-getting-started)
- [Hackathon Resources](#-hackathon-resources)

---

## 🛑 The Problem
Traditional anti-piracy solutions force 4K live broadcasts through third-party synchronous cloud APIs. This adds latency and risks catastrophic stream failure. When pirates strip metadata or record streams on their phones, standard watermarks fail.

**The SixPatrol Solution:** We decouple the process into an ultra-fast, on-premise **Edge SDK** (for zero-latency watermarking) and an asynchronous **Cloud Backend** (for heavy AI visual search and automated legal enforcement).

---

## 🧬 The 6-Layer Defense Architecture
SixPatrol embeds an indelible "DNA" into every frame, ensuring that no matter how the media is manipulated, ownership can be proven.

1. **Layer 1: C2PA Manifest** - Cryptographic metadata injecting the original broadcaster's signature.
2. **Layer 2: PixelSeal** - Invisible visual steganography applied directly to the video frames via LibTorch.
3. **Layer 3: AudioSeal** - Imperceptible acoustic watermarking applied to the audio channels.
4. **Layer 4: DINOv2** - Vision Transformer embeddings that understand semantic visual structure to detect heavily altered/cropped clips.
5. **Layer 5: DinoHash** - Perceptual hashing for rapid, sub-millisecond deduplication of exact matches.
6. **Layer 6: Immutable Ledger** - Cryptographic proof of ownership written to a Layer 2 blockchain.

---

## ✨ Core Features
* **The "Fail-Open" Edge SDK:** Built in C++, the edge node features a strict **15-millisecond Watchdog Timer**. If AI processing spikes, the SDK instantly bypasses the watermarking, guaranteeing the primary broadcast encoder never drops a frame.
* **Asynchronous Telemetry:** The SDK quietly pipes a lightweight 480p proxy stream to our Go/Gin backend without interrupting the primary 4K feed.
* **Gemini-Powered Enforcement:** When Python workers detect a piracy match in our Vector DB, the system triggers the **Google Gemini API** to autonomously draft and issue context-aware DMCA takedown notices.
* **HTMX Web Dashboard:** A blazingly fast, multi-tenant B2B portal offering real-time SSE (Server-Sent Events) piracy radar and billing metrics.

---

## 🛠 Tech Stack

### Edge Processing (Client Datacenter)
* **C++17** / **CMake** (vcpkg + FetchContent)
* **FFmpeg** (`libavcodec`, `libswscale`, `libswresample`)
* **LibTorch** (PyTorch C++ API)

### Cloud Infrastructure (Asynchronous)
* **Ingestion API:** Go (Gin framework)
* **Task Queue:** Redis & Celery
* **AI Workers:** Python (PyTorch, DINOv2)
* **Databases:** CockroachDB (Relational), Qdrant (Vector Search)
* **Google Cloud:** Compute Engine, Gemini API
* **Frontend:** HTMX, Tailwind CSS, HTML5

---

## 🏗 System Architecture
*The system is split into two primary boundaries: the synchronous client edge and the asynchronous SixPatrol cloud.*

1. **Ingestion:** Raw 4K feed enters the SixPatrol C++ SDK.
2. **Synchronous Split:** SDK applies L1-L3 watermarks in `<15ms` and passes the feed to the broadcast encoder.
3. **Asynchronous Split:** SDK pipes a proxy feed via HTTPS to the Go Ingestion API.
4. **AI Pipeline:** Python workers pull from Redis, generate DINOv2 embeddings, and store them in Qdrant.
5. **Enforcement:** Internet scrapes are matched against Qdrant. Matches trigger Gemini for DMCA drafting and register on the L6 Ledger.

---

## 🚀 Getting Started

### Prerequisites
* Linux (cxx11 ABI) or macOS
* CMake >= 3.20
* GCC/Clang with C++17 support
* Go 1.21+
* Python 3.10+

### Building the Edge SDK
1. Clone the repository:
   ```bash
   git clone [https://github.com/YourUsername/SixPatrol.git](https://github.com/YourUsername/SixPatrol.git)
   cd SixPatrol/sixpatrol-sdk
   ```
2. Download LibTorch (CPU version for Linux):
   ```bash
   wget [https://download.pytorch.org/libtorch/cpu/libtorch-cxx11-abi-shared-with-deps-2.2.0%2Bcpu.zip](https://download.pytorch.org/libtorch/cpu/libtorch-cxx11-abi-shared-with-deps-2.2.0%2Bcpu.zip)
   unzip libtorch-cxx11-abi-shared-with-deps-2.2.0+cpu.zip -d /path/to/libtorch
   ```
3. Configure and Build:
   ```bash
   cmake --preset local -DCMAKE_PREFIX_PATH=/path/to/libtorch
   cmake --build build
   ```

### Running the Cloud Backend
*(Assuming Redis and CockroachDB are running locally or via Docker)*
```bash
cd sixpatrol-cloud
go run main.go
```

---

## 🎥 Hackathon Resources
* **Demo Video:** [Insert YouTube Link Here]
* **Working Prototype:** [Insert Live URL Here]
* **Pitch Deck:** [Insert Link to Presentation]

---
*Built with ❤️ and AI for the 2026 Solution Challenge.*
