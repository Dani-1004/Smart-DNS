from sqlalchemy import create_engine, URL
from sqlalchemy.exc import SQLAlchemyError
from sqlalchemy.orm import sessionmaker, declarative_base
import os


DATABASE_URL = URL.create(
    "mysql+pymysql",
    username= os.environ['MYSQL_USER'],
    password= os.environ['MYSQL_PASSWORD'],  # plain (unescaped) text
    host="192.168.50.10",
    database="smartdns",
)

engine = create_engine(
    DATABASE_URL,
    pool_recycle=3600,
    echo=True,
)

SessionLocal = sessionmaker(bind=engine, autocommit=False, autoflush=False)

Base = declarative_base()
def create_table():
    Base.metadata.create_all(bind=engine)