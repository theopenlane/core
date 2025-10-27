#!/usr/bin/env python3

import os
import sys
from scim_client import SCIMClient

def provision_engineering_team(client: SCIMClient):
    """Provision a complete engineering team with users and groups"""

    print("Provisioning Engineering Team...")
    print("-" * 60)

    users = [
        {
            "username": "alice.engineer@example.com",
            "given_name": "Alice",
            "family_name": "Engineer",
            "display_name": "Alice Engineer"
        },
        {
            "username": "bob.developer@example.com",
            "given_name": "Bob",
            "family_name": "Developer",
            "display_name": "Bob Developer"
        },
        {
            "username": "carol.architect@example.com",
            "given_name": "Carol",
            "family_name": "Architect",
            "display_name": "Carol Architect"
        }
    ]

    created_users = []

    for user_data in users:
        print(f"Creating user: {user_data['username']}")
        user = client.create_user(
            username=user_data["username"],
            given_name=user_data["given_name"],
            family_name=user_data["family_name"],
            display_name=user_data.get("display_name"),
            active=True
        )
        created_users.append(user)
        print(f"  Created: {user['displayName']} (ID: {user['id']})")

    member_ids = [user['id'] for user in created_users]

    print("\nCreating Engineering group with all members...")
    eng_group = client.create_group(
        display_name="Engineering",
        members=member_ids
    )
    print(f"  Created: {eng_group['displayName']} (ID: {eng_group['id']})")
    print(f"  Members: {len(eng_group.get('members', []))}")

    print("\nCreating Backend subgroup...")
    backend_group = client.create_group(
        display_name="Backend Team",
        members=[created_users[0]['id'], created_users[1]['id']]
    )
    print(f"  Created: {backend_group['displayName']} (ID: {backend_group['id']})")

    print("\nCreating Frontend subgroup...")
    frontend_group = client.create_group(
        display_name="Frontend Team",
        members=[created_users[1]['id'], created_users[2]['id']]
    )
    print(f"  Created: {frontend_group['displayName']} (ID: {frontend_group['id']})")

    print("\n" + "=" * 60)
    print("Team Provisioning Summary:")
    print("=" * 60)
    print(f"Users created: {len(created_users)}")
    print(f"Groups created: 3 (Engineering, Backend Team, Frontend Team)")
    print("\nUser Details:")
    for user in created_users:
        print(f"  - {user['displayName']} ({user['userName']})")

    return {
        'users': created_users,
        'groups': [eng_group, backend_group, frontend_group]
    }


def cleanup_team(client: SCIMClient, team_data: dict):
    """Clean up all provisioned resources"""
    print("\n" + "=" * 60)
    print("Cleaning up resources...")
    print("-" * 60)

    for group in team_data['groups']:
        try:
            client.delete_group(group['id'])
            print(f"Deleted group: {group['displayName']}")
        except Exception as e:
            print(f"Error deleting group {group['displayName']}: {e}")

    for user in team_data['users']:
        try:
            client.delete_user(user['id'])
            print(f"Deleted user: {user['displayName']}")
        except Exception as e:
            print(f"Error deleting user {user['displayName']}: {e}")

    print("\nCleanup completed!")


def main():
    base_url = os.getenv('SCIM_BASE_URL')
    auth_token = os.getenv('SCIM_AUTH_TOKEN')

    if not base_url or not auth_token:
        print("Error: SCIM_BASE_URL and SCIM_AUTH_TOKEN environment variables must be set")
        sys.exit(1)

    client = SCIMClient(base_url, auth_token)

    try:
        team_data = provision_engineering_team(client)

        input("\nPress Enter to clean up resources...")
        cleanup_team(client, team_data)

    except Exception as e:
        print(f"\nError: {e}")
        import traceback
        traceback.print_exc()
        sys.exit(1)


if __name__ == '__main__':
    main()
