import requests
import json
import time
import re
import os

API_URL = "http://localhost:8080/vpp"
AUTH_TOKEN = "AexDQ4RyPi3jYETDHYFIxfFeQztzxBFoH3zZXGTTk0cI0RZqpzbqXM3epOeIOHik"

HEADERS = {
    "Authorization": f"Bearer {AUTH_TOKEN}",
    "Content-Type": "application/json"
}

# Use absolute path for vpp_dump.json to avoid FileNotFoundError when run via systemd
SCRIPT_DIR = os.path.dirname(os.path.abspath(__file__))
VPP_DUMP_JSON = os.path.join(SCRIPT_DIR, "vpp_dump.json")

def get_sw_ifindex_map():
    """Build map from interface name to sw_if_index."""
    resp = requests.get(f"{API_URL}/interfaces", headers=HEADERS)
    try:
        resp.raise_for_status()
    except Exception:
        print("Error get_sw_ifindex_map:", resp.text)
        raise
    data = resp.json()
    return {iface["name"]: iface["index"] for iface in data["interfaces"]}

def get_interface_map_by_name():
    """Return map: interface name -> interface dict (for host_if_name, etc)"""
    resp = requests.get(f"{API_URL}/interfaces", headers=HEADERS)
    try:
        resp.raise_for_status()
    except Exception:
        print("Error get_interface_map_by_name:", resp.text)
        raise
    data = resp.json()
    return {iface["name"]: iface for iface in data["interfaces"]}

def get_bond_id_from_name(name):
    m = re.match(r"BondEthernet(\d+)", name)
    return int(m.group(1)) if m else None

def create_bond(name, mtu):
    bond_id = get_bond_id_from_name(name)
    if bond_id is None:
        raise Exception(f"Invalid bond name: {name}")
    payload = {
        "mode": "lacp",  # Ubah jika ingin mode lain
        "id": bond_id
    }
    resp = requests.post(f"{API_URL}/bonds", json=payload, headers=HEADERS)
    try:
        resp.raise_for_status()
    except Exception:
        print(f"Bond create error for {name}:", resp.text)
        raise
    return resp.json().get("sw_if_index")

def create_vlan(parent_if_name, vlan_id, mtu, ip=None):
    sw_ifindex_map = get_sw_ifindex_map()
    parent_idx = sw_ifindex_map.get(parent_if_name)
    if parent_idx is None:
        raise Exception(f"Parent interface {parent_if_name} not found")
    payload = {
        "parent_if_index": parent_idx,
        "vlan_id": int(vlan_id),
        "mtu": mtu,
        "enable": True if mtu else False,
        "ip_address": ip or ""
    }
    resp = requests.post(f"{API_URL}/vlan/create", json=payload, headers=HEADERS)
    try:
        resp.raise_for_status()
    except Exception:
        print(f"VLAN create error for {parent_if_name}.{vlan_id}:", resp.text)
        raise
    return resp.json().get("sw_if_index")

def mirror_lcp(sw_if_index, host_if_name, host_if_type="tap"):
    """Call LCP mirror endpoint, idempotent."""
    url = f"{API_URL}/lcp/mirror"
    payload = {
        "sw_if_index": sw_if_index,
        "host_if_name": host_if_name,
        "host_if_type": host_if_type
    }
    resp = requests.post(url, json=payload, headers=HEADERS)
    try:
        resp.raise_for_status()
    except Exception:
        # If already exists or idempotent error, ignore
        if resp.status_code == 409 or (resp.status_code == 500 and "already exists" in resp.text):
            print(f"LCP mirror {host_if_name} already exists, skipping error.")
        else:
            print(f"LCP mirror error for {host_if_name} (sw_if_index={sw_if_index}):", resp.text)
            raise

def set_admin_state(sw_if_index, up):
    url = f"{API_URL}/interfaces/{sw_if_index}/{'enable' if up else 'disable'}"
    resp = requests.post(url, headers=HEADERS)
    try:
        resp.raise_for_status()
    except Exception:
        print(f"Set admin state error for {sw_if_index}:", resp.text)
        raise

def set_mtu(sw_if_index, mtu):
    url = f"{API_URL}/vlan/{sw_if_index}/mtu"
    payload = {"sw_if_index": sw_if_index, "mtu": mtu}
    resp = requests.post(url, json=payload, headers=HEADERS)
    try:
        resp.raise_for_status()
    except Exception:
        print(f"Set MTU error for {sw_if_index}:", resp.text)
        raise

def set_ip(sw_if_index, ip_addr):
    url = f"{API_URL}/vlan/{sw_if_index}/ip"
    payload = {"sw_if_index": sw_if_index, "ip_address": ip_addr}
    resp = requests.post(url, json=payload, headers=HEADERS)
    try:
        resp.raise_for_status()
    except Exception as e:
        # Jika error -105, abaikan (idempotent: IP sudah ada)
        if resp.status_code == 500 and '"details":-105' in resp.text:
            print(f"IP {ip_addr} already exists on {sw_if_index}, skipping error.")
        else:
            print(f"Set IP error for {sw_if_index}:", resp.text)
            raise

def wait_vpp_api(url=f"{API_URL}/interfaces", timeout=60):
    """Wait for VPP API to be ready before proceeding (for systemd integration)."""
    for _ in range(timeout):
        try:
            r = requests.get(url, headers=HEADERS)
            if r.status_code == 200:
                print("VPP API ready.")
                return True
        except Exception:
            pass
        time.sleep(1)
    raise Exception(f"VPP API not ready after {timeout} seconds")

def main():
    # Wait for VPP API to be ready (important for auto start via systemd)
    wait_vpp_api()

    with open(VPP_DUMP_JSON) as f:
        config = json.load(f)
    sw_ifindex_map = get_sw_ifindex_map()
    iface_by_name = get_interface_map_by_name()

    # Step 1: Create bonds, subinterfaces, and LCP mirror if needed
    for iface in config["interfaces"]:
        name = iface["name"]
        type_ = iface.get("type", "")
        mtu = iface.get("mtu", 0)
        host_if_name = iface.get("host_if_name", None)

        # Skip tap interface (host side of LCP), handled by LCP/mirror
        if name.startswith("tap"):
            continue

        # Bond
        if type_ == "unknown" and name.startswith("BondEthernet") and '.' not in name:
            if name not in sw_ifindex_map:
                print(f"Create bond: {name}")
                try:
                    create_bond(name, mtu)
                except Exception as e:
                    print(f"Failed to create bond {name}: {e}")
                time.sleep(1)
        # VLAN/Subinterface
        if '.' in name and name.startswith("BondEthernet"):
            parent, vlanid = name.split('.')
            if name not in sw_ifindex_map:
                print(f"Create VLAN: {name} (parent={parent}, vlan={vlanid})")
                ip = iface["ip_addresses"][0] if iface.get("ip_addresses") else None
                try:
                    create_vlan(parent, vlanid, mtu, ip)
                except Exception as e:
                    print(f"Failed to create VLAN {name}: {e}")
                time.sleep(1)
    # Refresh
    sw_ifindex_map = get_sw_ifindex_map()
    iface_by_name = get_interface_map_by_name()

    # Step 2: LCP mirror: lakukan jika field host_if_name tersedia dan belum termirror
    for iface in config["interfaces"]:
        name = iface["name"]
        host_if_name = iface.get("host_if_name", None)

        # Only mirror for VPP interface, not tap
        if name.startswith("tap"):
            continue

        if host_if_name:
            sw_if_index = sw_ifindex_map.get(name)
            if not sw_if_index:
                print(f"Cannot LCP mirror {name}: not found in VPP")
                continue
            # Check if already exists (host_if_name exists in iface_by_name)
            if host_if_name in iface_by_name:
                print(f"LCP {name} <-> {host_if_name} already exists, skip mirror.")
            else:
                print(f"LCP mirror {name} -> {host_if_name}")
                try:
                    mirror_lcp(sw_if_index, host_if_name, host_if_type="tap")
                except Exception as e:
                    print(f"Failed to LCP mirror {name} -> {host_if_name}: {e}")
                time.sleep(1)
    # Refresh again
    sw_ifindex_map = get_sw_ifindex_map()
    iface_by_name = get_interface_map_by_name()

    # Step 3: Set admin up/down, set MTU, set IP address (skip tap interface)
    for iface in config["interfaces"]:
        name = iface["name"]
        admin_up = iface.get("admin_up", False)
        mtu = iface.get("mtu", 0)
        ip_list = iface.get("ip_addresses", [])

        # Skip tap interface
        if name.startswith("tap"):
            print(f"Interface {name} is tap, skip set state/mtu/ip")
            continue

        sw_if_index = sw_ifindex_map.get(name)
        if sw_if_index is None:
            print(f"Interface {name} not found (skip set state/mtu/ip)")
            continue

        print(f"Set {name} admin {'up' if admin_up else 'down'}")
        try:
            set_admin_state(sw_if_index, admin_up)
        except Exception as e:
            print(f"Failed to set admin {name}: {e}")

        if mtu:
            print(f"Set {name} MTU {mtu}")
            try:
                set_mtu(sw_if_index, mtu)
            except Exception as e:
                print(f"Failed set MTU for {name}: {e}")

        for ip in ip_list:
            print(f"Set {name} IP {ip}")
            try:
                set_ip(sw_if_index, ip)
            except Exception as e:
                print(f"Failed set IP for {name}: {e}")

if __name__ == "__main__":
    main()
