# jce-codec
jce-codec 是 jce 序列化方案基础类型的编码实现，不包括 list、map、struct 等复杂结构，后者结构在代码生成中进行实现，也基于基础类型的序列化。代码生成的逻辑在 jce2go 项目中实现。

# jce
jce 是一种 TTLV 的二进制序列化方案，最早是在腾讯 rpc 框架 Tars 中实现的，也被称为 tars 协议，广泛使用于腾讯内部框架。

本项目基于 jce 序列化，与 jce 序列化大致一致，只是在一些细节上进行了优化，同时对代码进行重构，但总的来说和 jce 区别不大，为了表示对 jce 的敬意，故命名还是称之为 jce


# TTLV
TTLV 是 <Tag，Type，Length，Value> 四元组，用于对基础类型的序列化，即：
Tag：字段 id
Type：字段类型
Length：字段长度
Value：字段数据

这种编码方式也是二进制编码中最常见的一种，具体不再赘述，感兴趣的同学可以继续深入学习。

需要注意的是，对于有些基础类型，比如 int32，则 Type 就已经说明了 Length 的大小，故无需再写 Length 字段。
同时，Length 的长度具体又用多少字节来表示，这个也是一个问题。


# 序列化类型表
|值 | 	类型	       | 备注                       |
|:----:|:-------------:|---------------------------|
|0	 |   int1	       | 紧跟 1 个字节整型数据      |
|1	 |   int2	       | 紧跟 2 个字节整型数据      |
|2	 |   int4	       | 紧跟 4 个字节整型数据      |
|3	 |   int8	       | 紧跟 8 个字节整型数据      |
|4	 |   float4	       | 紧跟 4 个字节浮点型数据    |
|5	 |   float8	       | 紧跟 8 个字节浮点型数据    |
|6	 |   String	       | 紧跟长度字段，再跟内容|
|7	 |   Map	       | 紧跟长度字段，再跟 [key, value] 对列表|
|8	 |   List	       | 紧跟长度字段，再跟元素列表 |
|9	 |   自定义结构开始 | 自定义结构开始标志         |
|10	 |   自定义结构结束 | 自定义结构结束标志，Tag 为 0|
|11	 |   数字           | 0	表示数字 0，后面不跟数据|
|12	 |   SimpleList	   | 简单列表（目前用在 byte 数 组），紧跟一个类型字段（目前只支持 byte），紧跟一个整型数据表示长度，再跟 byte 数据|
|13	 |   -	           | -                          |
|14	 |   -	           | -                          |

这个类型表只用于序列化，和具体的语言无关，不同的语言最终都会转换的字节流中的 Type 都是这张表


# 序列化方案

## 基本结构

前面说了，数据的序列化分为 4 个部分，即 type、tag、length、data，而 type 和 tag 的范围都不大，通常：
1. tag 最大为 255 
2. type 上面给定的也只有 15 种

故，其实 tag 和 type 可以进行优化，放在一起存储，称为 head

而 length 字段有的类型没有，故总的方案如下：


```
|---head ----|---length----|---data--|
| type、tag  |  不一定有   |  data   |

```

## head 编码
现有的 type 只设计了 13 种，故 4b 就能表示，而 tag 很多时候也不会超过 16，故 4b 也能表示，故，很多时候其实 1B 就能表示 head。
故，根据 tag 的值大小，我们可以设计一个变长的编码，当 tag <15 用 1B 来表示，超过就用 2B

序列化 head，即 type+tag
方案如下：
1. 如果 tag < 15, 则编码为：
```
-------------------
| Type	 | Tag    |
| 4 bits | 4 bits |
-------------------
```

2. 如果 tag >= 15, 则编码为：
```
----------------------------
| Type	 | Tag 1  | Tag 2  |
| 4 bits | 4 bits | 1 byte |
----------------------------
```
其中 tag1 存默认值 15，真正的 tag 值存于 tag2 位置

为什么要像上面这样设计？而不是直接 type、tag 分别两个字节？
主要是考虑到 tag 很可能没有 15 大，只需 4bit 就能编码，而不用 8bit，同时 type 也 4bit 就能放下，那么
总的其实 1Byte 就能存，所以就根据 tag 的大小进行了位的压缩


## length 编码
高位为 0：用 7b 表示长度，范围为 0~127，length 一共 1B

```
---------------
|  length(1B) |
|  0xxx xxxx  |
---------------
```

高位为 1：用 31b 表示长度，范围为 0~2^31-1，length 一共 4B

```
------------------
|  length(4B)    |
|  0xxx xxxx  3B |
-----------------
```


## 数字(int1、int2、int4、int8、float4、float8)

方案如下：

```
|----------------|
| head  |  data  |
|----------------|
```

或者说
```
|----------------------|
| type  | tag |  data  |
|----------------------|
```

具体的 data 长度根据类型分配，即类型后那个数字就表示 length 字节个数，比如 int1 表示 1B


## 零(zero)
当数字类型等于 0 时，只需写入 Zero 类型即可，后面的 data 不用再写，用于优化为 0 的情况


```
|-------|
| head  |
|-------|
```


## 字符串(string)

```
|--------------------------------------------|
| head (1B or 2B) | length(1B or 4B) | data  |
|--------------------------------------------|
```

## simpleList

```
------------------------------------------
|  head | length | data type | byte list |
------------------------------------------
```

## list

```
---------------------------------------------
| head(type、tag) |   length  |     data    |
|    1 or 2 B     |     4B    |       ?     |
---------------------------------------------
```

data 可以是任何 type，直接递归其他类型的序列化即可，不过这里就有点问题，即内部的 data item 的 tag 其实是无用的，不过为了编码方便（直接递归），所以现在都是直接写为 0

## map

```
--------------------------------------------------------------
| head(type、tag) |   length  |     data (key、value list)   |
|    1 or 2 B     |     4B    |           ?                  |
--------------------------------------------------------------
```
类似于 list，只是 data 长度是两倍的 length，因为是 key、value 对

## struct
只是在 struct 的开始和结尾分别弄一个特殊的 type 即可

```
----------------------------------
| begin | ..............   | end |
----------------------------------
```

# 特定语言序列化
上面介绍了通用的序列化方案，在序列化特定语言时，还需要将这个语言的数据结构和上面的通用方案进行一个映射，然后才能进行序列化。

下面介绍 go 语言的序列化方案映射


| go 数据结构 | 	序列化方案	       | 
|:----:|:-------------:|
| bool	 |   int1	       |
|uint8	 |   int1	       |
|uint16 |   int2	       |
|uint32 |   int4	       |
|uint64	 |  int8 	       |
|int8	 |   int1	       |
|int16 |   int2	       |
|int32 |   int4	       |
|int64	 |   int8	       |
| byte | int1|
|rune | int 4| 
|float32	 |   float4	       |
|float64 |   float8	       |
|string	 |   string	       |
|map	 |   map	           |
|struct	 |   struct	       |
|slice	 |   list |
|arry	 |   list |
| []byte | simplieList |
| []int8 | simpleList | 
| []uint8 | simpleList | 
| comparable | 不支持 |
| any | 不支持 | 
| int | 不支持 | 
|uint	 |   不支持           |
|uintptr	 |   不支持	   | 
|complex64	 |   不支持	           | 
|complex128	 |   不支持           | 
| channel | 不支持 | 


# 优化设计
1. head 编码

head 存储 type、tag 时，使用边长编码优化设计

2. length 编码

根据统计，大量的 length 其实 1B 就能表示，故使用变长编码优化

3. zero 设计

当数字为 0 时，直接存 type 即可，后面的数据字段就优化掉了

4. 数字范围优化

当数字的大小比较小时，用更小的数据类型去存储

# 现有问题
1. list、map 内部数据的 tag 

~~默认都写为 0、 1 等，其实是无效数据，冗余只是为了方便编码，看以后是不是可以进行优化，即弄一套 head 只有 type 的，然后如果是 list 或 map 的内部数据，就用这个 head~~

这里 head 已经做了优化，tag < 125 那么和 type 就用了 1B，也不好继续优化了，毕竟 type 始终要使用一个字节的

2. float64 的优化

float64 的值范围在 float32 内时，不能优化为 float32 来存储，因为 IEEE754 编码会失真


# todolist
* [x] 编码方案的设计
* [x] 基础编码实现
* [x] 序列化测试
* [x] 性能优化
* [x] 压力测试
* [x] struct 内部时 struct 时的实现
* [x] struct 写 begin、end 字段
* [x] 长度字段的实现
* [ ] 重构代码，分层设计


# 参考

* [【后台开发拾遗】通信协议演进与 JCE 协议详解](https://blog.csdn.net/jiange_zh/article/details/86562232)
