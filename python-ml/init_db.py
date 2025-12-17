import mysql.connector

def ensure_database_exists():
    conn = mysql.connector.connect(
        host="localhost",
        user="x",
        password="x"
    )
    cursor = conn.cursor()
    cursor.execute("CREATE DATABASE IF NOT EXISTS dns_db")
    conn.commit()
    cursor.close()
    conn.close()
