# ConfigMap Intro

"node-resource-topo" ConfigMap defines the resource spec in the whole cluster.
Node-resource-manager on each node will perform create or update action according the resource definition in ConfigMap.

## Supported Node Resources

### LVM

- LVM VolumeGroup creation, according to the VG defined in ConfigMap;
- Create PV from local disk and add into VG;
- Don't support deletion or shrink of VG, to avoid data lost risk;

### QuotaPath

- QuotaPath creation and mount, according to the definition in ConfigMap;
- Don't support deletion or shrink of QuotaPath, to avoid data lost risk;

### PMEM

- Format pmem device as local memory, it can be later used by pod;

## How to define target node

We use the kubernetes label selector to choose target node:

```yaml
key: kubernetes.io/hostname
operator: In
value: xxxxx
```

- key: match the key in Node labels;
- operator: Labels selector operator,
  - In: matched only if current value equals to the value of the key in Node Labels;
  - NotIn: matched only if current value not equals to the value of key in the Node Labels;
  - Exists: matched if current key exists in Node Labels;
  - DoesNotExist: matched if current key not exists in Node Labels;

- value: match the corresponding value of the key in Node labels;

## Example

### LVM example

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

The above ConfigMap defines:

- select nodes with label `kubernetes.io/hostname: cn-zhangjiakou.192.168.3.114`, and create LVM VolumeGroup named `volumegroup1` from device `/dev/vdb` and `/dev/vdc`;
- select nodes without label `kubernetes.io/nodetype: localdisk`, and create LVM VolumeGroup named `volumegroup2` from all ecs cloud disks;
- select nodes with label `kubernetes.io/hostname: cn-beijing.192.168.3.35`, and create LVM VolumeGroup named `volumegroup1` from pmem in `region0`;

LVM currently supports three types of devices:

- `type: device` define lvm on top of local block devices, the volumegroup's name is specified in `name` field;
- `type: alibabacloud-local-disk` define lvm on top of attached cloud disks for alicloud ecs, the volumegroup's name is specified in `name` field;
- `type: pmem` define lvm on top of local pmem resources, the volumegroup's name is specified in `name` field, the pmem regions is specified in `regions` field;

### QuotaPath example

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

The above ConfigMap defines:

- select nodes with label `kubernetes.io/hostname: cn-beijing.192.168.3.35`, and create quota path `/mnt/path1` mounted from device `/dev/vdb`, and format to ext4 filesystem with prjquota option;
- select nodes with label `kubernetes.io/hostname: cn-beijing.192.168.3.36`, and create quota path `/mnt/path2` mounted from pmem in `region0`, and format to ext4 filesystem with prjquota, shared option;

QuotaPath currently supports two types of devices:

- `type: device` define quota path on top of local block device, the quota path is specified in `name` field:
  - options: mount options, `prjquota` is mandatory;
  - fstype: filesystem type, ext4 is used by default;
  - devices: block device to be mounted, you should specify only one device;
- `type: pmem` define quota path on top of local pmem resources, the quota path is specified in `name` field, you can speficy pmem regions in `regions` field;

### PMEM example

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

The above ConfigMap defines:

- select nodes with label `kubernetes.io/hostname: cn-beijing.192.168.3.37`, and create memory from pmem in `region0`;

PMEM only support `type: pmem`, you can speficy pmem regions in `regions` field, name field is only a symbol and has no actual usage.
