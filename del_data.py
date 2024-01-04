import redis
import pymysql  #pip3 install mysqlclient
import time



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


while True:
   try:
      zset_keys = []
      r = redis.Redis(connection_pool=pool)
      db = pymysql.connect(host=mysql_host,user= mysql_user, password=mysql_passwd, database=mysql_database)

      cursor = db.cursor()
      sql = """select ticks from inscription_info"""
      cursor.execute(sql)

      # 执行SQL语句
      cursor.execute(sql)
      # 获取所有记录列表
      results = cursor.fetchall()
      for row in results:
         tick = row[0]
         zset_keys.append(zset_key_pre + tick)
      


      db.close()


      for zset_key in zset_keys:
         # r.zremrangebyrank(zset_key, 10, -1)
         members_after_1000th = r.zrevrange(zset_key, 1000, -1, withscores=True)

         #  # 删除排名在第 1000 名以后的成员
         if members_after_1000th:
            last_member_score = members_after_1000th[0][1]  # 获取第一个成员的分数
            last_member_end = members_after_1000th[-1][1]  # 获取最后一个成员的分数
            r.zremrangebyscore(zset_key, last_member_end, last_member_score)  # 删除分数在最后一个成员分数之后的所有成员
   
   except:
         print("Error: unable to fetch data")

   time.sleep(60)