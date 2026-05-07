# Media Alignment Observatory Indonesia

This project transparently monitors **narrative alignment** across Indonesian mass media using data-driven methods.

**Goal:**
- Measure how closely media narratives align with official government sources
- Detect temporal coordination patterns (when multiple outlets publish highly similar content in a short time window)
- Provide objective, evidence-based insights to improve public media literacy in Indonesia

**Status:** Pilot Phase

## Current Features
- RSS + full article scraping
- Indonesian text preprocessing
- Sentence embeddings & cosine similarity analysis
- Data storage in SQLite
- Basic visualization

## Methodology (High-Level)
We combine:
- Traditional NLP techniques
- Modern sentence embeddings (Indonesian-optimized models)
- Temporal synchronization analysis

Future phases will include LLM-based framing & sentiment analysis.

## Quick Start

### 1. Installation
```bash
git clone https://github.com/axeltjs/media-alignment-observatory.git
cd media-alignment-observatory

python -m venv venv
source venv/bin/activate          # On Windows: venv\Scripts\activate

pip install -r requirements.txt
