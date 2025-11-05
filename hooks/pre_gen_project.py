#!/usr/bin/env python3
"""
Pre-generation hook for cookiecutter template.
Checks if required tools are installed before generating the project.
"""

import subprocess
import sys
from typing import List, Tuple


def check_command_exists(command: str) -> Tuple[bool, str]:
    """
    Check if a command exists and is executable.

    Args:
        command: The command to check

    Returns:
        Tuple of (exists: bool, version_info: str)
    """
    try:
        # Use 'which' to check if command exists
        result = subprocess.run(
            ["which", command], capture_output=True, text=True, check=True
        )

        # If command exists, try to get version information
        try:
            if command == "go":
                version_result = subprocess.run(
                    [command, "version"], capture_output=True, text=True, check=True
                )
            elif command == "pulumi":
                version_result = subprocess.run(
                    [command, "version"], capture_output=True, text=True, check=True
                )
            elif command == "talosctl":
                version_result = subprocess.run(
                    [command, "version", "--client", "--short"],
                    capture_output=True,
                    text=True,
                    check=True,
                )
            else:
                version_result = subprocess.run(
                    [command, "--version"], capture_output=True, text=True, check=True
                )

            version_info = (
                version_result.stdout.strip() or version_result.stderr.strip()
            )

            # For talosctl, we want the second line which contains the actual version
            if command == "talosctl":
                lines = version_info.split("\n")
                if len(lines) > 1:
                    return True, lines[1].strip()
                else:
                    return True, lines[0].strip()
            else:
                return True, version_info.split("\n")[0]  # Return first line only

        except subprocess.CalledProcessError:
            # Command exists but version check failed
            return True, "version unknown"

    except subprocess.CalledProcessError:
        return False, ""


def check_required_tools() -> List[str]:
    """
    Check if all required tools are installed.

    Returns:
        List of missing tools
    """
    required_tools = ["go", "pulumi", "talosctl"]
    missing_tools = []

    print("Checking required tools...")
    print("=" * 50)

    for tool in required_tools:
        exists, version_info = check_command_exists(tool)

        if exists:
            print(f"✅ {tool}: {version_info}")
        else:
            print(f"❌ {tool}: NOT FOUND")
            missing_tools.append(tool)

    print("=" * 50)
    return missing_tools


def print_installation_instructions(missing_tools: List[str]) -> None:
    """
    Print installation instructions for missing tools.

    Args:
        missing_tools: List of missing tool names
    """
    print(
        "\n⚠️  Missing tools detected. Consider installing them for full functionality:"
    )
    print("=" * 60)

    installation_guides = {
        "go": {
            "description": "Go programming language",
            "macos": "brew install go",
            "linux": "Visit https://go.dev/doc/install",
            "url": "https://go.dev/doc/install",
        },
        "pulumi": {
            "description": "Pulumi Infrastructure as Code tool",
            "macos": "brew install pulumi/tap/pulumi",
            "linux": "curl -fsSL https://get.pulumi.com | sh",
            "url": "https://www.pulumi.com/docs/install/",
        },
        "talosctl": {
            "description": "Talos Linux CLI tool",
            "macos": "brew install siderolabs/tap/talosctl",
            "linux": "Visit https://www.talos.dev/v1.11/talos-guides/install/talosctl/",
            "url": "https://www.talos.dev/v1.11/talos-guides/install/talosctl/",
        },
    }

    for tool in missing_tools:
        if tool in installation_guides:
            guide = installation_guides[tool]
            print(f"\n📦 {tool} - {guide['description']}")
            print(f"   macOS:  {guide['macos']}")
            print(f"   Linux:  {guide['linux']}")
            print(f"   Docs:   {guide['url']}")

    print("\n" + "=" * 60)
    print("You can install these tools later and use the generated project.")
    print("Note: Some functionality may be limited without these tools.")


def main():
    """Main function to check prerequisites."""
    try:
        missing_tools = check_required_tools()

        if missing_tools:
            print_installation_instructions(missing_tools)
            print(
                f"\n⚠️  Template generation proceeding with {len(missing_tools)} tool(s) missing"
            )
        else:
            print(
                "\n✅ All required tools are installed. Proceeding with template generation..."
            )

    except Exception as e:
        print(f"\n⚠️  Warning: Error checking prerequisites: {e}")
        print("Proceeding with template generation anyway...")


if __name__ == "__main__":
    main()
