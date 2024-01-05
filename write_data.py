import redis
import pymysql  #pip3 install mysqlclient
import time
import json



# aias:trade:history:set:xxx

zset_key_pre = "aias:trade:history:set:"


redis_host = ''
redis_port = 6379
redis_passwd = 'tom'

mysql_host = ""
mysql_user = "aias"
mysql_passwd = "aias.yyds"
mysql_database = "aias"

pool = redis.ConnectionPool(host=redis_host, port=redis_port, decode_responses=True, password=redis_passwd, db=12)

db = pymysql.connect(host=mysql_host,user= mysql_user, password=mysql_passwd, database=mysql_database)

cursor = db.cursor()
sql = """select ticks,from_address,to_address,hash,amount,time,status,number from trade_history order by number desc limit 1000"""
cursor.execute(sql)

# 执行SQL语句
cursor.execute(sql)
# 获取所有记录列表
results = cursor.fetchall()
for row in results:
    data = {}
    data["ticks"] = row[0]
    data["fromAddress"] = row[1]
    data["toAddress"] = row[2]
    data["hash"] = row[3]
    data["amount"] = row[4]
    data["time"] = row[5]
    data["status"] = row[6]
    data["number"] = row[7]
    if data["fromAddress"] == data["toAddress"]:
        data["method"] = "mint"
    else:
        data["method"] = "transfer"


    key = "aias:trade:history:set:aias"
    source = float(data["method"])
    value = json.dumps(data)
    print(value)
    # ret = r.zadd(key,{value:source})



db.close()


