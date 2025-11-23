import copy


def extract_docker_tag(image_ref: str) -> str:
    """Extract the Docker tag from an image reference."""
    last_colon_index = image_ref.rfind(":")

    if last_colon_index != -1:
        return image_ref[last_colon_index + 1 :]
    else:
        return "latest"


def get_longest_common_substring(key: str, strs: list[str]) -> str:
    """Find the longest common substring among a list of strings, including the key."""
    if not strs:
        return ""

    newStrs = copy.deepcopy(strs)
    newStrs.insert(0, key)

    shortest_str = min(newStrs, key=len)
    n = len(shortest_str)

    for length in range(n, 0, -1):
        for i in range(n - length + 1):
            substring = shortest_str[i : i + length]
            if all(substring in s for s in newStrs):
                return substring

    return ""


def parse_image_address(image_address) -> dict[str, str | None]:
    """Parse a Docker image address into its components."""
    parts: dict[str, str | None] = {
        "registry": None,
        "namespace": None,
        "repository": None,
        "tag": None,
        "digest": None,
    }

    if "@" in image_address:
        address_part, parts["digest"] = image_address.split("@", 1)
    else:
        address_part = image_address

    if ":" in address_part and "/" not in address_part.split(":")[-1]:
        base_part, parts["tag"] = address_part.rsplit(":", 1)
    else:
        base_part = address_part
        parts["tag"] = "latest"

    path_parts = base_part.split("/")

    if len(path_parts) == 1:
        parts["image_name"] = path_parts[0]
    elif len(path_parts) == 2:
        if "." in path_parts[0] or ":" in path_parts[0]:
            parts["registry"] = path_parts[0]
            parts["image_name"] = path_parts[1]
        else:
            parts["namespace"] = path_parts[0]
            parts["image_name"] = path_parts[1]
    elif len(path_parts) >= 3:
        parts["registry"] = path_parts[0]
        parts["namespace"] = path_parts[1]
        parts["image_name"] = "/".join(path_parts[2:])

    return parts
