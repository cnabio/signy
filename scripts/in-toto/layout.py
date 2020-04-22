# Imports.

# 2nd-party.
from typing import Any, Dict, List, Optional

# 3rd-party.
from in_toto.models.layout import Layout

# Artifact rules
# https://github.com/in-toto/docs/blob/master/in-toto-spec.md#433-artifact-rules

ArtifactRule = List[str]
ArtifactRules = List[ArtifactRule]

def allow(pattern: str) -> ArtifactRule:
    return ["ALLOW", pattern]

def create(pattern: str) -> ArtifactRule:
    return ["CREATE", pattern]

def delete(pattern: str) -> ArtifactRule:
    return ["DELETE", pattern]


def disallow(pattern: str) -> ArtifactRule:
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

def match_materials(pattern: str, step_name: str, source_path_prefix: Optional[str] = None, destination_path_prefix: Optional[str] = None) -> ArtifactRule:
    return _match(pattern, "MATERIALS", step_name, source_path_prefix, destination_path_prefix)

def match_products(pattern: str, step_name: str, source_path_prefix: Optional[str] = None, destination_path_prefix: Optional[str] = None) -> ArtifactRule:
    return _match(pattern, "PRODUCTS", step_name, source_path_prefix, destination_path_prefix)

def modify(pattern: str) -> ArtifactRule:
    return ["MODIFY", pattern]

def require(pattern: str) -> ArtifactRule:
    return ["REQUIRE", pattern]

# Steps
# https://github.com/in-toto/docs/blob/master/in-toto-spec.md#431-steps

Command = List[str]
# https://github.com/python/typing/issues/182#issuecomment-185996450
PublicKeys = Dict[str, Any]
Step = Dict[str, Any]
Steps = List[Step]

def step(name: str, materials: ArtifactRules, products: ArtifactRules, pubkeys: PublicKeys, expected_command: Command = [], threshold: int = 1) -> Step:
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

# Inspections
# https://github.com/in-toto/docs/blob/master/in-toto-spec.md#432-inspections

Inspection = Dict[str, Any]
Inspections = List[Inspection]

def inspection(name: str, materials: ArtifactRules, products: ArtifactRules, command: Command = []) -> Inspection:
    return {
        "_name": name,
        "expected_materials": materials,
        "expected_products": products,
        "run": command
    }

# Layouts
# https://github.com/in-toto/docs/blob/master/in-toto-spec.md#43-file-formats-layout

def layout(steps: Steps, inspections: Inspections, keys: PublicKeys, expires_days: int = 0, expires_months: int = 0, expires_years: int = 0) -> Layout:
    l = Layout.read({
        "_type": "layout",
        "keys": keys,
        "steps": steps,
        "inspect": inspections
    })
    # Set expiration timestamp.
    l.set_relative_expiration(days=expires_days, months=expires_months, years=expires_years)
    return l