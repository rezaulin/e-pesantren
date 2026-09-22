import mysql.connector

try:
    conn = mysql.connector.connect(user='root', password='', host='127.0.0.1', database='pesantren_multi')
    cursor = conn.cursor()
    cursor.execute("SHOW TABLES")
    for (table_name,) in cursor:
        print(table_name)
    conn.close()
except Exception as e:
    print("Error:", e)
