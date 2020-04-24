#!/usr/bin/env python3

# Imports.

# 1st-party.
from keys import (
    KeyDict, Threshold,
    get_private_keys_from_keyring,
    get_public_keys_from_keyring,
    sorted_list_of_keyids,
    write_and_read_new_keys,
)
from layout import (
    Layout, Step,
    allow, layout, match_materials, match_products, modify, step,
)

# 3rd-party
from in_toto.models.metadata import Metablock

# Constants
# We assume using JsonPath to select elements within a bundle.
# TODO: How should we escape JsonPath to work around fnmatch?
BUNDLE_ROOT = 'file://bundle.json$'
BUNDLE_ALL = f'{BUNDLE_ROOT}..*'

#  Step functions.

def get_pubkeys(this_step_name, m: int = 1, n: int = 1) -> KeyDict:
    threshold = Threshold(m, n)
    keyring = write_and_read_new_keys(this_step_name, threshold)
    return get_public_keys_from_keyring(keyring)

def get_developer_step(m: int = 1, n: int = 1) -> (Step, KeyDict):
    this_step_name = 'developer'
    this_step_pubkeys = get_pubkeys(this_step_name, m, n)
    this_step = step(
        name = this_step_name,
        # Developers do NOT receive any materials/input.
        # However, they MUST produce the following products/output. 
        products = [
            # 1. Developers are allowed to write anything under the bundle.
            allow(BUNDLE_ALL),
            # TODO: 2. Developers are NOT allowed to fill in image digests and sizes?
        ],
        pubkeys = sorted_list_of_keyids(this_step_pubkeys),
        threshold = m,
    )
    return this_step, this_step_pubkeys

def get_machine_step(developer_step: Step, m: int = 1, n: int = 1) -> (Step, KeyDict):
    this_step_name = 'machine'
    this_step_pubkeys = get_pubkeys(this_step_name, m, n)
    this_step = step(
        name = this_step_name,
        # Machines MUST receive the following materials.
        materials = [
            # 1. Machines MUST receive the same product developers produced.
            match_products(
                pattern = BUNDLE_ALL,
                # FIXME: Should be {developer_step.name}.
                step_name = developer_step["name"],
            ),
        ],
        # Machines MUST produce the following products.
        products = [
            # 1. Machines MUST produce the same product developers produced.
            match_materials(
                pattern = BUNDLE_ALL,
                step_name = this_step_name,
            ),
            # 2. However, machines MUST fill in ONLY image digests and sizes.
            # TODO: double-check pattern.
            modify(f'{BUNDLE_ROOT}.images.*'),
        ],
        pubkeys = sorted_list_of_keyids(this_step_pubkeys),
    )
    return this_step, this_step_pubkeys

# Layout functions.

def get_layout(m: int = 1, n: int = 1, expires_years: int = 1) -> Metablock:
    developer_step, developer_pubkeys = get_developer_step()
    machine_step, machine_pubkeys = get_machine_step(developer_step)

    threshold = Threshold(m, n)
    keyring = write_and_read_new_keys('layout', threshold)
    signed = layout(
        steps = [developer_step, machine_step],
        keys = {**developer_pubkeys, **machine_pubkeys},
        expires_years = expires_years,
    )

    metablock = Metablock(signed=signed)
    for k in get_private_keys_from_keyring(keyring).values():
        metablock.sign(k)
    print(str(metablock))
    return metablock

# CLI.

if __name__ == '__main__':
    get_layout()
