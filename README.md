# tabgo

游戏打表工具

* 基本类型bool,string,int,float
* 输出json,lua
* 数组(支持多维数组，结构体数组)
* 结构体定义(支持嵌套结构体,数组成员)
* 服务端客户端分别打表(标记为:client的字段服务端表将会忽略)


![Alt text](20221125102046.png)

### string

对于string类型，填写值的时候无需使用""包裹


#### 嵌套的string值

tabgo支持string作为数组或结构体的成员。

但是，当string作为内嵌成员时，其值必须用""包裹。

例如：

对于类型string[]

值[hello,world]是非法的，正确的值应该是["hello","world"]

如果在值内包含了字符"需要使用\\"转义。











