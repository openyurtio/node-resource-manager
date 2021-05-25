# ConfigMap 介绍

"node-resource-topo" ConfigMap 定义整个集群的设备拓扑，每个节点的 node-resource-manager 会根据 ConfigMap 中的资源定义决定是否创建或者更新资源；

## 节点资源支持

### LVM
- 根据 ConfigMap 中定义，创建 LVM VolumeGroup；
- 在 VG 中添加、扩展 PV；根据 ConfigMap 中定义的 VG 信息，判断是否执行 VG 扩容命令；
- 考虑数据安全性，不支持 VG 的删除、缩容操作；

### QuotaPath
- QuotaPath 的创建, 根据 ConfigMap 中的定义来初始化相关本地资源设备以 QuotaPath 的形式挂载到指定路径上；
- 不支持相关 QuotaPath 的变更, 删除等操作；

### PMEM
- 支持将持久化内存设备初始化为内存格式。后续可以直接被挂载到pod内部目录中使用；

## 如何定义节点

我们通过如下三个 key/value 来共同定义资源所在的节点：

```yaml
key: kubernetes.io/hostname
operator: In
value: xxxxx
```
- key: 匹配 Kubernetes Node Labels 中的 key 的值；
- operator: Kubernetes 定义的 Labels selector operator，主要包含如下四种操作符；
    - In: 只有 value 的值与 Node 上 Labels key 对应的值相同的时候才会匹配；
    - NotIn: 只有 value 的值与 Node 上 Labels key 对应的 value 的值 ***不*** 相同的时候才会匹配；
    - Exists: 只要 Node 的 Labels 上存在 Key 就会匹配；
    - DoesNotExist: 只要 Node 的 Labels 上 ***不*** 存在 Key 就会匹配；
- value: 匹配 Kubernetes Node Labels 的 key 对应的 value 的值；

## 常用模板用例及说明

### LVM 

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: node-resource-topo
  namespace: kube-system
data:
  volumegroup: |-
    volumegroup:
    - name: volumegroup1
      key: kubernetes.io/hostname
      operator: In
      value: cn-zhangjiakou.192.168.3.114
      topology:
        type: device
        devices:
        - /dev/vdb
        - /dev/vdc

    - name: volumegroup2
      key: kubernetes.io/nodetype
      operator: NotIn
      value: localdisk
      topology:
        type: alibabacloud-local-disk

    - name: volumegroup1
      key: kubernetes.io/hostname
      operator: Exists
      value: cn-beijing.192.168.3.35
      topology:
        type: pmem
        regions:
        - region0
```
上面的 ConfigMap 定义了
1. 在拥有 Label key 等于 kubernetes.io/hostname 并且 Label key value 等于 cn-zhangjiakou.192.168.3.114 的 Node 上创建一个名字为 volumegroup1 的 LVM VolumeGroup, 这个 VolumeGroup 由宿主机上的 ```/dev/vdb, /dev/vdc``` 两个块设备组成
2. 在拥有 Label key 等于 kubernetes.io/hostname 并且 Label key value 不等于 localdisk 的 Node 上创建一个名字为 volumegroup2 的 LVM VolumeGroup，这个 LVM VolumeGroup 由宿主机上的所有本地盘组成
3. 在拥有 Label key 等于 kubernetes.io/hostname 的 Node 上创建一个名字为 volumegroup1 的 LVM VolumeGroup, 这个 VolumeGroup 由宿主机上的 Pmem 设备 region0 组成

lvm目前仅支持三种定义资源拓扑的方式
- 当定义 ```type: device``` 的时候是通过 nrm 所在宿主机上存在的块设备 devices 进行 lvm 的声明，声明的块设备会组成一个 volumegroup, volumegroup 的名字由 name 字段指定，供后续应用启动时分配 logical volume。```type: device``` 类型与下面的 ```devices``` 字段绑定；
- 当定义 ```type: alibabacloud-local-disk``` 的时候指的是使用 urm 所在宿主机上所有的本地盘 (选择 ecs 类型带有本地盘的 instance , 例如 本地 SSD 型 i2, 手动挂载到 ecs 上的云盘不是本地盘) 共同创建一个 名称为 name 值的 volumegroup；
- 当定义 ```type: pmem``` 的时候是使用 urm 所在宿主机上的 pmem 资源创建一个名称为 name 值的 volumegroup, 其中 regions 可以指定当前机器上多个 pmem region 资源。```type: pmem``` 类型与下面的 ```regions``` 字段绑定。

### QuotaPath

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: node-resource-topo
  namespace: kube-system
data:
  quotapath: |-
    quotapath:
    - name: /mnt/path1
      key: kubernetes.io/hostname
      operator: In
      value: cn-beijing.192.168.3.35
      topology:
        type: device
        options: prjquota
        fstype: ext4
        devices:
        - /dev/vdb

    - name: /mnt/path2
      key: kubernetes.io/hostname
      operator: In
      value: cn-beijing.192.168.3.36
      topology:
        type: pmem
        options: prjquota,shared
        fstype: ext4
        regions: 
        - region0
```
上面的 ConfigMap 定义了
1. 在拥有 Label key 等于 kubernetes.io/hostname 并且 Label key value 等于 cn-beijing.192.168.3.35 的 Node 上的 ```/mnt/path1``` 上以 prjquota 类型挂载 ```/dev/vdb``` 的块设备， 并且格式化成 project quota ext4 格式。
2. 在拥有 Label key 等于 kubernetes.io/hostname 并且 Label key value 等于 cn-beijing.192.168.3.36 的 Node 上的 ```/mnt/path2``` 上以prjquota 类型挂载 ```region0``` 的 Pmem 设备， 并且格式化成 project quota ext4 格式。

QuotaPath 类型的本地资源仅支持两种定义资源拓扑的方式：
- 当定义 ```type: device``` 的时候，使用的是 nrm 所在宿主机的块设备进行 QuotaPath 的初始化，初始化路径是 name 字段定义的值, 下面解释下其他几个字段的定义：
    - options: 块设备在被挂载的时候使用的参数。无特殊需求使用例子中提供的参数即可；
    - fstype: 格式化块设备使用的文件系统，默认使用 ext4；
    - devices：挂载使用的块设备，只可以声明一个；

### pmem

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: node-resource-topo
  namespace: kube-system
data:
  memory: |-
    memory:
    - name: test1
      key: kubernetes.io/hostname
      operator: In
      value: cn-beijing.192.168.3.37
      topology:
        type: pmem
        regions: 
        - region0
```

上面的 ConfigMap 定义了
1. 在拥有 Label key 等于 kubernetes.io/hostname 并且 Label key value 等于 cn-beijing.192.168.3.37 的 Node 上用 Pmem 的 Region0 设备初始化一个内存格式的 Pmem 设备。

持久化内存类型本地资源只支持 pmem type 的挂载，其中 regions 代表需要被格式化成 memory 的 pmem 设备，name 字段在这里只起到标识作用，无实际意义。
