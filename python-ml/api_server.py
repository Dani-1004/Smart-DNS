# api_server.py
import os
import asyncio
from datetime import datetime, timedelta
from fastapi import FastAPI, HTTPException
from pydantic import BaseModel
import redis
from typing import Dict

# Import your existing functions / models
# - load_models (already loads tfidf_vectorizer, nb_model, lstm_model, tokenizer)
# - analyze_domain(domain) that returns (category, action) - blocking
# - DNSDatabase class with update_domain_category(domain, category, action) and connect/get_unknown_domains
# Adjust import paths as needed.
from dns_blocker import load_models, analyze_domain  # replace with actual module name
from dns_data import DNSDatabase
from sqlalchemy.orm import Session
from database import create_table, SessionLocal
from models import DomainClassification

# ========== CONFIG ==========
REDIS_HOST = os.getenv("REDIS_HOST", "192.168.50.15")
REDIS_PORT = int(os.getenv("REDIS_PORT", "6379"))
REDIS_DB = int(os.getenv("REDIS_DB", "0"))
REDIS_TTL = int(os.getenv("REDIS_TTL", 24 * 3600))  # 24 hours 
REDIS_PASSWORD = os.getenv("REDIS_PASSWORD", "x")
QUEUE_SIZE = int(os.getenv("QUEUE_SIZE", 1000))

app = FastAPI(title="ML Domain Analyzer API")

# Initialize Redis (synchronous client)
redis_client = redis.Redis(
    host=REDIS_HOST, 
    port=REDIS_PORT, 
    db=REDIS_DB,
    password=REDIS_PASSWORD, 
    decode_responses=True
    )

# ===== Queue Bucket System =====
queue = asyncio.Queue(maxsize=QUEUE_SIZE)

# Load ML models once at startup (blocking; do it when app starts)
tfidf_vectorizer, nb_model, lstm_model, tokenizer = load_models()

# Pydantic request model
class DomainRequest(BaseModel):
    domain: str
    record_type: str
    record_value: str | None = None
    ttl: int | None = None

# Utility: normalize domain for DB/Redis keys
def normalize_domain(d: str) -> str:
    d = d.strip().lower()
    if not d:
        return d
    # remove trailing dot for storage keys (optional: keep appended dot in DB if you expect it)
    return d.rstrip('.')

# ================================
# Background Worker
# ================================
async def queue_worker():
    print("[WORKER] Worker started (sequential mode)")

    while True:
        job = await queue.get()

        print(f"[WORKER] Processing: {job['domain']}")

        # Process blocking (NO async, NO threads)
        process_and_store(
            job["domain"],
            job["rValue"],
            job["rType"],
            job["rTTL"]
        )

        queue.task_done()

        print(f"[WORKER] Done. Remaining in queue: {queue.qsize()}")



@app.on_event("startup")
async def startup_event():
    print("[INIT] Creating tables")
    create_table()

    print("[INIT] Starting worker")
    asyncio.create_task(queue_worker())


def process_and_store(domain: str, rValue, rType, rTTL):
    """
    Sequential blocking job. NO threads, NO concurrency.
    """
    domain_norm = normalize_domain(domain)
    if not domain_norm:
        return

    try:
        category, action = analyze_domain(domain_norm)

        print("\n==================== ML RESULT ====================")
        print("Domain    :", domain_norm)
        print("Category  :", category)
        print("Action    :", action)
        print("====================================================\n")

        # Update SQL
        database = DNSDatabase()
        db: Session = next(database.get_db())
        record = db.query(DomainClassification).filter_by(domain=domain).first()

        if record:
            record.category = category
            record.action = action
            record.updated_at = datetime.utcnow()
            db.commit()
            print("[DB] Updated")
        else:
            new_rec = DomainClassification(
                domain=domain,
                category=category,
                action=action,
                updated_at=datetime.utcnow()
            )
            db.add(new_rec)
            db.commit()
            print("[DB] Created")

        # Update Redis
        redis_client.hset(
            f"class:{domain}",
            mapping={
                "category": category,
                "action": action
            }
        )
        print("[REDIS] Updated")

        return {"status": "ok"}

    except Exception as e:
        print(f"[ERROR] process_and_store({domain_norm}): {e}")
        return {"status": "error", "message": str(e)}



# === Endpoint to accept domain and queue processing ===
@app.post("/api/domain", status_code=202)
async def receive_domain(payload: DomainRequest):
    domain = payload.domain

    if not domain:
        raise HTTPException(400, "domain required")

    # --- Redis Cache Check ---
    redis_key = f"domain:{domain}"
    if redis_client.exists(redis_key):
        cached = redis_client.hgetall(redis_key)
        return {
            "status": "cached",
            "domain": domain,
            "category": cached.get("category"),
            "action": cached.get("action")
        }

    # --- DB Recent Check ---
    db = SessionLocal()
    db_record = db.query(DomainClassification)\
                  .filter(DomainClassification.domain == domain)\
                  .first()

    if db_record:
        now = datetime.utcnow()
        if now - db_record.updated_at < timedelta(hours=1):
            return {
                "status": "skipped",
                "message": "Recently analyzed, skipped"
            }

    # --- Queue Insert ---
    try:
        await queue.put({
            "domain": domain,
            "rValue": payload.record_value,
            "rType": payload.record_type,
            "rTTL": payload.ttl
        })
    except asyncio.QueueFull:
        raise HTTPException(503, "Server busy — queue full")

    return {
        "status": "queued",
        "message": "Domain added to processing queue"
    }
