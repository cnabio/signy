#!/usr/bin/env python3

# Imports.

# 1st-party.
from keys import KeyDict
from layout import (
    CREATE, MATCH_MATERIALS, MATCH_PRODUCTS, MODIFY,
    Step,
    layout, step,
)

# 3rd-party.
import click

# Constants
# NOTE: we use JsonPath to record leaves,
# but NOT to specify patterns in artifact rules.
# We use fnmatch to specify patterns in artifact rules.
BUNDLE_ROOT = 'file://bundle.json$'
BUNDLE_ALL = f'{BUNDLE_ROOT}.*'

#  Step functions.

def get_developer_step(trust_dir: str) -> (Step, KeyDict):
    return step(
        trust_dir,
        'developer',
        # Developers do NOT receive any materials/input.
        # However, they MUST produce the following products/output. 
        products = [
            # 1. Developers are allowed to create anything under the bundle.
            CREATE(BUNDLE_ALL),
        ],
    )

def get_machine_step(trust_dir: str, developer_step: Step, this_step_name: str = 'machine') -> (Step, KeyDict):
    return step(
        trust_dir,
        this_step_name,
        # Machines MUST receive the following materials.
        materials = [
            # 1. Machines MUST receive the same product developers produced.
            MATCH_PRODUCTS(
                pattern = BUNDLE_ALL,
                # FIXME: Should be {developer_step.name}.
                step_name = developer_step['name'],
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
            MODIFY(f'{BUNDLE_ROOT}.images.*'),
        ],
    )

# Layout functions.

@click.command()
@click.option('-d', '--dir', 'trust_dir', default='~/.signy', help='Directory where the trust data is persisted to.')
@click.option('-o', '--output', 'output_filename', default='minimal.root.layout', help='Filename to write root layout to.')
def write_layout(trust_dir: str, output_filename: str) -> None:
    developer_step, developer_pubkeys = get_developer_step(trust_dir)
    machine_step, machine_pubkeys = get_machine_step(trust_dir, developer_step)

    metablock = layout(
        trust_dir,
        'minimal-root-layout',
        steps = [developer_step, machine_step],
        keys = {**developer_pubkeys, **machine_pubkeys},
        expires_years = 1,
    )

    metablock.dump(output_filename)
    print(f'Wrote minimal root layout to: {output_filename}')
    print(f'jq -C "." {output_filename} | less -R')

# CLI.

if __name__ == '__main__':
    write_layout() # pylint: disable=no-value-for-parameter