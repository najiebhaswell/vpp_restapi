
### /vpp/bonds

#### GET
##### Summary:

List Bond Interfaces

##### Description:

List all bond interfaces.

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | object |
| 500 | Internal Server Error | object |

#### POST
##### Summary:

Create Bond Interface

##### Description:

Create a new bond interface with members.

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| body | body | Bond Config {mode: string, interfaces: []int} | Yes | object |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 201 | Created | object |
| 400 | Bad Request | object |
| 500 | Internal Server Error | object |

### /vpp/bonds/{sw_if_index}

#### DELETE
##### Summary:

Delete Bond Interface

##### Description:

Delete a bond interface by sw_if_index.

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| sw_if_index | path | Bond Interface Index | Yes | integer |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | object |
| 400 | Bad Request | object |
| 500 | Internal Server Error | object |

### /vpp/bonds/{sw_if_index}/member

#### POST
##### Summary:

Add Bond Member

##### Description:

Add a member to a bond interface.

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| sw_if_index | path | Bond Interface Index | Yes | integer |
| body | body | Member {member_index: int} | Yes | object |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | object |
| 400 | Bad Request | object |
| 500 | Internal Server Error | object |

### /vpp/interfaces

#### GET
##### Summary:

List all interfaces

##### Description:

Get all VPP interfaces with status.

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | object |
| 500 | Internal Server Error | object |

### /vpp/interfaces/loopback

#### POST
##### Summary:

Create Loopback Interface

##### Description:

Create a new loopback interface.

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| body | body | Loopback Config {mac_address: string} | Yes | object |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 201 | Created | object |
| 400 | Bad Request | object |
| 500 | Internal Server Error | object |

### /vpp/interfaces/{sw_if_index}

#### DELETE
##### Summary:

Delete Interface

##### Description:

Delete a loopback or bond interface by sw_if_index.

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| sw_if_index | path | Interface Index | Yes | integer |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | object |
| 400 | Bad Request | object |
| 500 | Internal Server Error | object |

### /vpp/interfaces/{sw_if_index}/disable

#### POST
##### Summary:

Disable interface

##### Description:

Set interface to admin down.

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| sw_if_index | path | Interface Index | Yes | integer |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | object |
| 400 | Bad Request | object |
| 500 | Internal Server Error | object |

### /vpp/interfaces/{sw_if_index}/enable

#### POST
##### Summary:

Enable interface

##### Description:

Set interface to admin up.

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| sw_if_index | path | Interface Index | Yes | integer |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | object |
| 400 | Bad Request | object |
| 500 | Internal Server Error | object |

### /vpp/lcp/mirror

#### POST
##### Summary:

Mirror VPP Interface to Host

##### Description:

Mirror a VPP interface (LCP pair) to the host's kernel namespace

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| body | body | Mirror Config {sw_if_index: int, host_if_name: string, host_if_type: string, netns: string} | Yes | object |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | object |
| 400 | Bad Request | object |
| 500 | Internal Server Error | object |

### /vpp/version

#### GET
##### Summary:

Show VPP version

##### Description:

Get the running VPP version and build date

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | object |
| 500 | Internal Server Error | object |

### /vpp/vlan/create

#### POST
##### Summary:

Create VLAN Subinterface

##### Description:

Create a VLAN subinterface with options.

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| body | body | VLAN Config | Yes | [vlan.VLANCreateRequest](#vlan.vlancreaterequest) |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 201 | Created | object |
| 400 | Bad Request | object |
| 500 | Internal Server Error | object |

### /vpp/vlan/{sw_if_index}

#### DELETE
##### Summary:

Delete VLAN Subinterface

##### Description:

Delete a VLAN subinterface by sw_if_index.

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| sw_if_index | path | VLAN SwIfIndex | Yes | integer |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | object |
| 400 | Bad Request | object |
| 500 | Internal Server Error | object |

### /vpp/vlan/{sw_if_index}/disable

#### POST
##### Summary:

Enable VLAN

##### Description:

Enable a VLAN interface by sw_if_index.

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| sw_if_index | path | VLAN SwIfIndex | Yes | integer |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | object |
| 400 | Bad Request | object |
| 500 | Internal Server Error | object |

### /vpp/vlan/{sw_if_index}/enable

#### POST
##### Summary:

Enable VLAN

##### Description:

Enable a VLAN interface by sw_if_index.

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| sw_if_index | path | VLAN SwIfIndex | Yes | integer |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | object |
| 400 | Bad Request | object |
| 500 | Internal Server Error | object |

### /vpp/vlan/{sw_if_index}/ip

#### POST
##### Summary:

Set VLAN IP Address

##### Description:

Set the IP address of a VLAN interface.

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| sw_if_index | path | VLAN SwIfIndex | Yes | integer |
| body | body | IP Request | Yes | [vlan.VLANActionRequest](#vlan.vlanactionrequest) |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | object |
| 400 | Bad Request | object |
| 500 | Internal Server Error | object |

### /vpp/vlan/{sw_if_index}/mtu

#### POST
##### Summary:

Set VLAN MTU

##### Description:

Set the MTU of a VLAN interface.

##### Parameters

| Name | Located in | Description | Required | Schema |
| ---- | ---------- | ----------- | -------- | ---- |
| sw_if_index | path | VLAN SwIfIndex | Yes | integer |
| body | body | MTU Request | Yes | [vlan.VLANActionRequest](#vlan.vlanactionrequest) |

##### Responses

| Code | Description | Schema |
| ---- | ----------- | ------ |
| 200 | OK | object |
| 400 | Bad Request | object |
| 500 | Internal Server Error | object |

### Models


#### vlan.VLANActionRequest

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| ip_address | string |  | No |
| mtu | integer |  | No |
| sw_if_index | integer |  | Yes |

#### vlan.VLANCreateRequest

| Name | Type | Description | Required |
| ---- | ---- | ----------- | -------- |
| enable | boolean |  | No |
| ip_address | string |  | No |
| mtu | integer |  | No |
| parent_if_index | integer |  | Yes |
| vlan_id | integer |  | Yes |