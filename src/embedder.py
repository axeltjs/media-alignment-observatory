# src/embedder.py
from sentence_transformers import SentenceTransformer
import numpy as np
from pathlib import Path
import json
import pandas as pd
from sklearn.metrics.pairwise import cosine_similarity

class Embedder:
    def __init__(self, model_name: str = "LazarusNLP/Indo-Sentence-BERT"):
        print(f"Loading model: {model_name}")
        self.model = SentenceTransformer(model_name)
        print("Model loaded successfully.")
    
    def embed_text(self, text: str):
        return self.model.encode(text, convert_to_numpy=True, show_progress_bar=False)
    
    def embed_batch(self, texts: list[str]):
        return self.model.encode(texts, batch_size=32, convert_to_numpy=True, show_progress_bar=True)
    
    def compute_similarity(self, emb1, emb2):
        return cosine_similarity([emb1], [emb2])[0][0]
    
    def process_articles(self, data_dir: Path = Path("data/raw")):
        articles = []
        for json_file in data_dir.glob("*.json"):
            with open(json_file, "r", encoding="utf-8") as f:
                data = json.load(f)
                articles.append({
                    "id": data["id"],
                    "media": data["media"],
                    "title": data["title"],
                    "text": data["text"][:2000],  # Limit token
                    "published": data["published"]
                })
        
        df = pd.DataFrame(articles)
        print(f"Computing embeddings for {len(df)} articles...")
        
        embeddings = self.embed_batch(df["text"].tolist())
        df["embedding"] = list(embeddings)
        
        return df

# Contoh penggunaan
if __name__ == "__main__":
    embedder = Embedder()
    df = embedder.process_articles()
    print(df[["media", "title"]].head())
