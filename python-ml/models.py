from sqlalchemy import Column, Integer, String, DateTime
from database import Base
from datetime import datetime


# class DNSRecord(Base):
#     __tablename__ = "dns_records"

#     id = Column(Integer, primary_key=True, index=True)
#     domain = Column(String(255), index=True)
#     record_type = Column(String(20))
#     record_value = Column(String(255))
#     ttl = Column(Integer)
#     category = Column(String(50), default="unknown")
#     action = Column(String(50), default="forward")
#     created_at = Column(DateTime, default=datetime.utcnow)

class DomainClassification(Base):
    __tablename__ = "dns_class"
    id = Column(Integer, primary_key=True, index=True)
    domain = Column(String(255), index=True)
    category = Column(String(50), default="unknown")
    action = Column(String(50), default="forward")
    updated_at = Column(DateTime, default=datetime.utcnow)    