# src/scraper.py
import feedparser
import newspaper
from datetime import datetime
import json
from pathlib import Path
import hashlib

DATA_DIR = Path("data/raw")
DATA_DIR.mkdir(parents=True, exist_ok=True)

def generate_article_id(url: str) -> str:
    return hashlib.md5(url.encode()).hexdigest()[:12]

def scrape_rss(media_name: str, rss_url: str, max_articles: int = 20):
    print(f"Scraping {media_name}...")
    feed = feedparser.parse(rss_url)
    
    articles = []
    for entry in feed.entries[:max_articles]:
        try:
            article = newspaper.Article(entry.link)
            article.download()
            article.parse()
            
            if len(article.text) < 200:  # Skip jika terlalu pendek
                continue
                
            article_data = {
                "id": generate_article_id(entry.link),
                "media": media_name,
                "title": article.title,
                "url": entry.link,
                "published": entry.get("published", str(datetime.now())),
                "text": article.text,
                "authors": article.authors,
                "scraped_at": datetime.now().isoformat()
            }
            
            # Save individual file
            file_path = DATA_DIR / f"{article_data['id']}.json"
            with open(file_path, "w", encoding="utf-8") as f:
                json.dump(article_data, f, ensure_ascii=False, indent=2)
            
            articles.append(article_data)
            print(f"  ✓ {article.title[:80]}...")
            
        except Exception as e:
            print(f"  ✗ Error scraping {entry.link}: {e}")
    
    return articles

# Contoh penggunaan
if __name__ == "__main__":
    test_media = {
        "Kompas": "https://www.kompas.com/rss",
        # Tambah media lain nanti
    }
    
    for name, url in test_media.items():
        scrape_rss(name, url)
