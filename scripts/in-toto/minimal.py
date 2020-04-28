#!/usr/bin/env python3

# Imports.

# 1st-party.
from keys import (
    KeyDict, Threshold,
    get_private_keys_from_keyring,
)
from layout import (
    ALLOW, MATCH_MATERIALS, MATCH_PRODUCTS, MODIFY,
    Step,
    get_step, layout, write_and_read_new_keys,
)

# 3rd-party
from in_toto.models.metadata import Metablock

# Constants
# We assume using JsonPath to select elements within a bundle.
# TODO: How should we escape JsonPath to work around fnmatch?
BUNDLE_ROOT = 'file://bundle.json$'
BUNDLE_ALL = f'{BUNDLE_ROOT}..*'

#  Step functions.

def get_developer_step() -> (Step, KeyDict):
    return get_step(
        'developer',
        # Developers do NOT receive any materials/input.
        # However, they MUST produce the following products/output. 
        products = [
            # 1. Developers are allowed to write anything under the bundle.
            ALLOW(BUNDLE_ALL),
            # TODO: 2. Developers are NOT allowed to fill in image digests and sizes?
        ],
    )

def get_machine_step(developer_step: Step, this_step_name: str = 'machine') -> (Step, KeyDict):
    return get_step(
        this_step_name,
        # Machines MUST receive the following materials.
        materials = [
            # 1. Machines MUST receive the same product developers produced.
            MATCH_PRODUCTS(
                pattern = BUNDLE_ALL,
                # FIXME: Should be {developer_step.name}.
                step_name = developer_step["name"],
            ),
        ],
        # Machines MUST produce the following products.
        products = [
            # 1. Machines MUST produce the same product developers produced.
            MATCH_MATERIALS(
                pattern = BUNDLE_ALL,
                step_name = this_step_name,
            ),
            # 2. However, machines MUST fill in ONLY image digests and sizes.
            # TODO: double-check pattern.
            MODIFY(f'{BUNDLE_ROOT}.images.*'),
        ],
    )

# Layout functions.

def get_layout() -> Metablock:
    developer_step, developer_pubkeys = get_developer_step()
    machine_step, machine_pubkeys = get_machine_step(developer_step)

    threshold = Threshold(1, 1)
    keyring = write_and_read_new_keys('layout', threshold)
    signed = layout(
        steps = [developer_step, machine_step],
        keys = {**developer_pubkeys, **machine_pubkeys},
        expires_years = 1,
    )
    metablock = Metablock(signed=signed)
    for k in get_private_keys_from_keyring(keyring).values():
        metablock.sign(k)
    print(str(metablock))
    return metablock

# CLI.

if __name__ == '__main__':
    get_layout()
