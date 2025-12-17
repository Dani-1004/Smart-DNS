from init_db import ensure_database_exists
from sqlalchemy import create_engine, URL
from sqlalchemy.orm import sessionmaker, declarative_base
from sqlalchemy.exc import SQLAlchemyError
from models import DomainClassification
from sqlalchemy import create_engine
import os

DATABASE_URL = URL.create(
    "mysql+pymysql",
    username=os.getenv("MYSQL_USER", "root"),
    password=os.getenv("MYSQL_PASSWORD", "root"),  # plain (unescaped) text
    host=os.getenv("MYSQL_HOST", "localhost"),
    database=os.getenv("MYSQL_DB", "smartdns"),
)

class DNSDatabase:
    def __init__(self):
        # Create SQLAlchemy engine & session factory
        self.engine = create_engine(
            DATABASE_URL,
            echo=False,
            pool_pre_ping=True,
            pool_recycle=3600,
        )
        self.SessionLocal = sessionmaker(
            autocommit=False,
            autoflush=False,
            bind=self.engine
        )

        # Ensure tables exist
        self.Base = declarative_base()
        
        self.Base.metadata.create_all(bind=self.engine)

    # ------------------------------------------------------
    # Create new DB session
    # ------------------------------------------------------
    def get_db(self):
        db = self.SessionLocal()
        try:
            yield db
        finally:
            db.close()

    
    # ------------------------------------------------------
    # Select domains where category = 'unknown'
    # ------------------------------------------------------
    def get_unknown_domains(self, limit=20):
        db = next(self.get_db())
        try:
            rows = db.query(DomainClassification).filter(
                DomainClassification.category == "unknown"
            ).limit(limit).all()

            return [r.domain.strip().lower() for r in rows]

        except SQLAlchemyError as e:
            print("[SQL ERROR] get_unknown_domains:", e)
            return []

    # ------------------------------------------------------
    # Update category & action for a domain
    # ------------------------------------------------------
    def update_domain_category(self, domain, category, action):
        db = next(self.get_db())
        try:
            record = db.query(DomainClassification).filter(
                DomainClassification.domain == domain
            ).first()

            if not record:
                print(f"[SQL WARN] Domain not found: {domain}")
                return False

            record.category = category
            record.action = action

            db.commit()
            print(f"[SQL] Updated: {domain} → {category} ({action})")
            return True

        except SQLAlchemyError as e:
            db.rollback()
            print("[SQL ERROR] update_domain_category:", e)
            return False

    # ------------------------------------------------------
    # Insert DNS record (used from Golang → FastAPI)
    # ------------------------------------------------------
    def insert_record(self, domain):
        db = next(self.get_db())
        try:
            rec = DomainClassification(
                domain=domain
            )
            db.add(rec)
            db.commit()
            db.refresh(rec)
            return rec.id

        except SQLAlchemyError as e:
            db.rollback()
            print("[SQL ERROR] insert_record:", e)
            return None