# HostCollision
基于golang编写的多线程Host爆破/扫描/碰撞工具

## 使用方法

```shell
./main -i <ip_file_path> -d <host_file_path> -n <numuber_of_goroutine>  -o <output_file_path> -s 1000
```

example:

```shell
./main -i ip.txt -d host.txt -o output.txt -s 1500 -n 15 -m 20 -r 80
```

## 参数说明

带 * 为必须

-i *ip文件  
-o *输出文件位置  
-d *host文件  
-n 协程数量，可省略，默认为20   
-s sleep 时间，单位为ms，默认是1000  
-r 相似率，默认为85  
-m 最大Host阈值，单个ip的host成功数量超过这阈值不会再进行爆破(通常是有问题的), 默认为50

## 文件格式

ip.txt
```
10.1.10.2
172.10.1.3
```

host.txt
```
www.aaa.com
www.bbb.com
```

## 截图
![Demo][./image/demo.png]
