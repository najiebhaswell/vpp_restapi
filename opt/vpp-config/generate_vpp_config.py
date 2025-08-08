import requests
import json
import os

API_URL = "http://localhost:8080/vpp"
AUTH_TOKEN = "AexDQ4RyPi3jYETDHYFIxfFeQztzxBFoH3zZXGTTk0cI0RZqpzbqXM3epOeIOHik"
OUTPUT_PATH = "/opt/vpp-config/vpp_dump.json"

HEADERS = {
    "Authorization": f"Bearer {AUTH_TOKEN}",
    "Content-Type": "application/json"
}

def fetch_interfaces():
    resp = requests.get(f"{API_URL}/interfaces", headers=HEADERS)
    resp.raise_for_status()
    return resp.json().get("interfaces", [])

def fetch_bonds():
    resp = requests.get(f"{API_URL}/bonds", headers=HEADERS)
    resp.raise_for_status()
    return resp.json().get("bonds", [])

def build_vpp_config():
    interfaces = fetch_interfaces()
    bonds = fetch_bonds()
    config = {"interfaces": []}
    # Buat map: bond name -> anggota
    bond_members = {bond['name']: bond.get('members', []) for bond in bonds}
    for iface in interfaces:
        entry = {
            "admin_up": iface.get("admin_up", False),
            "index": iface.get("index", -1),
            "ip_addresses": iface.get("ip_addresses", []),
            "link_up": iface.get("link_up", False),
            "mtu": iface.get("mtu", 0),
            "name": iface.get("name", ""),
            "type": iface.get("type", "unknown")
        }
        if "host_if_name" in iface:
            entry["host_if_name"] = iface["host_if_name"]
        # Jika bond, tambahkan member jika ada
        if iface["name"] in bond_members:
            entry["members"] = bond_members[iface["name"]]
        config["interfaces"].append(entry)
    return config

def is_default_config(config):
    interfaces = config.get("interfaces", [])
    if len(interfaces) == 1 and interfaces[0]["name"] == "local0":
        return True
    for iface in interfaces:
        if iface["name"] != "local0":
            if iface.get("admin_up", False):
                return False
            if iface.get("ip_addresses"):
                return False
    return True

def main():
    config = build_vpp_config()
    if is_default_config(config):
        print("VPP config is still default, not saving.")
        return
    os.makedirs(os.path.dirname(OUTPUT_PATH), exist_ok=True)
    with open(OUTPUT_PATH, "w") as f:
        json.dump(config, f, indent=2)
    print(f"VPP config dumped to {OUTPUT_PATH}")

if __name__ == "__main__":
    main()
