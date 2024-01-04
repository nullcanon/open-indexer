import sys
import json
import threading,time
from threading import Thread
import redis
import websocket

import rel


# {
# 'ticks': 'aias', 
# 'fromAddress': '0x79f3c10d803458d63a92a1056ec9d6b8322a3547', 
# 'toAddress': '0x79f3c10d803458d63a92a1056ec9d6b8322a3547', 
# 'hash': '0x6f66a1bda6425262cd68aa87430827e9d840ee91c08030e8081c40e5186094a1', 
# 'amount': '1000000', 
# 'time': 1703284342, 
# 'status': '1', 
# 'number': 1060120
# }

# aias:trade:history:set:xxx
# xxx为铭文的名称

data = {"jsonrpc":"2.0","method":"tick_subscribe","params":["history"], "id":1}
json_data = json.dumps(data)


pool = redis.ConnectionPool(host='43.139.3.138', port=6379, decode_responses=True, password='tom', db=12)
r = redis.Redis(connection_pool=pool)

times = 0

def on_message(ws, message):
    jmessage = json.loads(message)
    history_list = jmessage['params']['result']
    if history_list is None:
        return
    for history in history_list:

        key = "aias:trade:history:set:" + history['ticks']
        source = history['number']
        value = str(history)
        ret = r.zadd(key,{value:source})
        print(ret,source)



def on_error(ws, error):
    print(error)


def on_close(ws, close_status_code, close_msg):
    print ("Retry : %s" % time.ctime())
    time.sleep(10)
    connect_websocket() # retry per 10 seconds


def on_open(ws):
    ws.send(json_data)
    # def run(*args):
    #     ws.send(json_data)
    #     print("Thread terminating...")

    # Thread(target=run).start()

def connect_websocket():
    websocket.enableTrace(True)
    if len(sys.argv) < 2:
        host = "ws://43.139.191.120:9999"
    else:
        host = sys.argv[1]
    ws = websocket.WebSocketApp(
        host, on_message=on_message, on_error=on_error, on_close=on_close
    )
    ws.on_open = on_open

    wst = threading.Thread(target=ws.run_forever())
    wst.daemon = True
    wst.start()

def re_connect_websocket():
    while True:  # 通过一个无限循环保持连接
        try:
            connect_websocket()
        except Exception as e:
            print(f"连接发生异常：{e}")
            times = times + 1
            if times > 10:
                sys.exit(1)
            time.sleep(30)  # 等待一段时间后重新连接


if __name__ == "__main__":
    connect_websocket()
