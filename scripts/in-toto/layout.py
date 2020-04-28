# Imports.

# 1st-party.
from keys import (
    KeyDict, Threshold,
    get_public_keys_from_keyring,
    sorted_list_of_keyids,
    write_and_read_new_keys,
)


# 2nd-party.
from typing import Any, Dict, List, Optional

# 3rd-party.
from in_toto.models.layout import Layout

# Artifact rules
# https://github.com/in-toto/docs/blob/master/in-toto-spec.md#433-artifact-rules

ArtifactRule = List[str]
ArtifactRules = List[ArtifactRule]

def ALLOW(pattern: str) -> ArtifactRule:
    return ["ALLOW", pattern]

def CREATE(pattern: str) -> ArtifactRule:
    return ["CREATE", pattern]

def DELETE(pattern: str) -> ArtifactRule:
    return ["DELETE", pattern]

def DISALLOW(pattern: str) -> ArtifactRule:
    return ["DISALLOW", pattern]

def _match(pattern: str, materials_or_products: str, step_name: str, source_path_prefix: Optional[str] = None, destination_path_prefix: Optional[str] = None) -> ArtifactRule:
    l = ["MATCH", pattern]

    if source_path_prefix:
        l += ["IN", source_path_prefix]

    l += ["WITH", materials_or_products]

    if destination_path_prefix:
        l += ["IN", destination_path_prefix]
    
    l += ["FROM", step_name]
    return l

def MATCH_MATERIALS(pattern: str, step_name: str, source_path_prefix: Optional[str] = None, destination_path_prefix: Optional[str] = None) -> ArtifactRule:
    return _match(pattern, "MATERIALS", step_name, source_path_prefix, destination_path_prefix)

def MATCH_PRODUCTS(pattern: str, step_name: str, source_path_prefix: Optional[str] = None, destination_path_prefix: Optional[str] = None) -> ArtifactRule:
    return _match(pattern, "PRODUCTS", step_name, source_path_prefix, destination_path_prefix)

def MODIFY(pattern: str) -> ArtifactRule:
    return ["MODIFY", pattern]

def REQUIRE(pattern: str) -> ArtifactRule:
    return ["REQUIRE", pattern]

# Steps
# https://github.com/in-toto/docs/blob/master/in-toto-spec.md#431-steps

Command = List[str]
Step = Dict[str, Any]
Steps = List[Step]

def step(name: str, materials: ArtifactRules = [], products: ArtifactRules = [], pubkeys: KeyDict = {}, expected_command: Command = [], threshold: int = 1) -> Step:
    assert threshold > 0, f'{threshold} <= 0'
    return {
        "_type": "step",
        "name": name,
        "expected_materials": materials,
        "expected_products": products,
        "pubkeys": pubkeys,
        # NOTE: There is no point in using this feature, because: (1) we do not
        # wrap in-toto around expected commands, and (2) in-toto cannot really
        # check them without a trusted OS.
        "expected_command": expected_command,
        "threshold": threshold,
    }

def get_pubkeys(this_step_name, m: int = 1, n: int = 1) -> KeyDict:
    threshold = Threshold(m, n)
    keyring = write_and_read_new_keys(this_step_name, threshold)
    return get_public_keys_from_keyring(keyring)

def get_step(this_step_name: str, m: int = 1, n: int = 1, materials: ArtifactRules = [], products: ArtifactRules = []) -> (Step, KeyDict):
    this_step_pubkeys = get_pubkeys(this_step_name, m, n)
    this_step = step(
        name = this_step_name,
        materials = materials,
        products = products,
        pubkeys = sorted_list_of_keyids(this_step_pubkeys),
        threshold = m,
    )
    return this_step, this_step_pubkeys

# Inspections
# https://github.com/in-toto/docs/blob/master/in-toto-spec.md#432-inspections

Inspection = Dict[str, Any]
Inspections = List[Inspection]

def inspection(name: str, materials: ArtifactRules = [], products: ArtifactRules = [], command: Command = []) -> Inspection:
    return {
        "_name": name,
        "expected_materials": materials,
        "expected_products": products,
        "run": command
    }

# Layouts
# https://github.com/in-toto/docs/blob/master/in-toto-spec.md#43-file-formats-layout

def layout(steps: Steps = [], inspections: Inspections = [], keys: KeyDict = {}, expires_days: int = 0, expires_months: int = 0, expires_years: int = 0) -> Layout:
    l = Layout.read({
        "_type": "layout",
        "keys": keys,
        "steps": steps,
        "inspect": inspections
    })
    # Set expiration timestamp.
    l.set_relative_expiration(days=expires_days, months=expires_months, years=expires_years)
    return l
