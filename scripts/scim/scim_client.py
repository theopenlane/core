#!/usr/bin/env python3

import os
import sys
import json
import requests
from typing import Dict, Any, Optional, List

class SCIMClient:
    def __init__(self, base_url: str, auth_token: str):
        self.base_url = base_url.rstrip('/')
        self.headers = {
            'Authorization': f'Bearer {auth_token}',
            'Content-Type': 'application/scim+json',
            'Accept': 'application/scim+json'
        }

    def create_user(self, username: str, given_name: str, family_name: str,
                   display_name: Optional[str] = None, active: bool = True) -> Dict[str, Any]:
        """Create a new SCIM user"""
        if not display_name:
            display_name = f"{given_name} {family_name}".strip()

        payload = {
            "schemas": ["urn:ietf:params:scim:schemas:core:2.0:User"],
            "userName": username,
            "name": {
                "givenName": given_name,
                "familyName": family_name
            },
            "displayName": display_name,
            "active": active
        }

        response = requests.post(
            f"{self.base_url}/v2/Users",
            headers=self.headers,
            json=payload
        )
        response.raise_for_status()
        return response.json()

    def get_user(self, user_id: str) -> Dict[str, Any]:
        """Get a SCIM user by ID"""
        response = requests.get(
            f"{self.base_url}/v2/Users/{user_id}",
            headers=self.headers
        )
        response.raise_for_status()
        return response.json()

    def list_users(self, start_index: int = 1, count: int = 100, filter_str: Optional[str] = None) -> Dict[str, Any]:
        """List SCIM users"""
        params = {
            'startIndex': start_index,
            'count': count
        }
        if filter_str:
            params['filter'] = filter_str

        response = requests.get(
            f"{self.base_url}/v2/Users",
            headers=self.headers,
            params=params
        )
        response.raise_for_status()
        return response.json()

    def find_user_by_username(self, username: str) -> Optional[Dict[str, Any]]:
        """Find a user by username"""
        users = self.list_users(filter_str=f'userName eq "{username}"')
        resources = users.get('Resources', [])
        return resources[0] if resources else None

    def update_user(self, user_id: str, operations: List[Dict[str, Any]]) -> Dict[str, Any]:
        """Update a SCIM user using PATCH"""
        payload = {
            "schemas": ["urn:ietf:params:scim:api:messages:2.0:PatchOp"],
            "Operations": operations
        }

        response = requests.patch(
            f"{self.base_url}/v2/Users/{user_id}",
            headers=self.headers,
            json=payload
        )
        response.raise_for_status()
        return response.json()

    def deactivate_user(self, user_id: str) -> Dict[str, Any]:
        """Deactivate a SCIM user"""
        return self.update_user(user_id, [
            {
                "op": "replace",
                "value": {"active": False}
            }
        ])

    def activate_user(self, user_id: str) -> Dict[str, Any]:
        """Activate a SCIM user"""
        return self.update_user(user_id, [
            {
                "op": "replace",
                "value": {"active": True}
            }
        ])

    def update_user_name(self, user_id: str, given_name: str, family_name: str) -> Dict[str, Any]:
        """Update a user's name using PATCH"""
        return self.update_user(user_id, [
            {
                "op": "replace",
                "path": "name.givenName",
                "value": given_name
            },
            {
                "op": "replace",
                "path": "name.familyName",
                "value": family_name
            }
        ])

    def delete_user(self, user_id: str) -> None:
        """Delete a SCIM user"""
        response = requests.delete(
            f"{self.base_url}/v2/Users/{user_id}",
            headers=self.headers
        )
        response.raise_for_status()

    def create_group(self, display_name: str, members: Optional[List[str]] = None) -> Dict[str, Any]:
        """Create a new SCIM group"""
        payload = {
            "schemas": ["urn:ietf:params:scim:schemas:core:2.0:Group"],
            "displayName": display_name
        }

        if members:
            payload["members"] = [{"value": member_id} for member_id in members]

        response = requests.post(
            f"{self.base_url}/v2/Groups",
            headers=self.headers,
            json=payload
        )
        response.raise_for_status()
        return response.json()

    def get_group(self, group_id: str) -> Dict[str, Any]:
        """Get a SCIM group by ID"""
        response = requests.get(
            f"{self.base_url}/v2/Groups/{group_id}",
            headers=self.headers
        )
        response.raise_for_status()
        return response.json()

    def list_groups(self, start_index: int = 1, count: int = 100, filter_str: Optional[str] = None) -> Dict[str, Any]:
        """List SCIM groups"""
        params = {
            'startIndex': start_index,
            'count': count
        }
        if filter_str:
            params['filter'] = filter_str

        response = requests.get(
            f"{self.base_url}/v2/Groups",
            headers=self.headers,
            params=params
        )
        response.raise_for_status()
        return response.json()

    def find_group_by_name(self, display_name: str) -> Optional[Dict[str, Any]]:
        """Find a group by display name"""
        groups = self.list_groups(filter_str=f'displayName eq "{display_name}"')
        resources = groups.get('Resources', [])
        return resources[0] if resources else None

    def add_group_member(self, group_id: str, user_id: str) -> Dict[str, Any]:
        """Add a member to a SCIM group"""
        payload = {
            "schemas": ["urn:ietf:params:scim:api:messages:2.0:PatchOp"],
            "Operations": [
                {
                    "op": "add",
                    "path": "members",
                    "value": [{"value": user_id}]
                }
            ]
        }

        response = requests.patch(
            f"{self.base_url}/v2/Groups/{group_id}",
            headers=self.headers,
            json=payload
        )
        response.raise_for_status()
        return response.json()

    def remove_group_member(self, group_id: str, user_id: str) -> Dict[str, Any]:
        """Remove a member from a SCIM group"""
        payload = {
            "schemas": ["urn:ietf:params:scim:api:messages:2.0:PatchOp"],
            "Operations": [
                {
                    "op": "remove",
                    "path": "members",
                    "value": [{"value": user_id}]
                }
            ]
        }

        response = requests.patch(
            f"{self.base_url}/v2/Groups/{group_id}",
            headers=self.headers,
            json=payload
        )
        response.raise_for_status()
        return response.json()

    def update_group_name(self, group_id: str, display_name: str) -> Dict[str, Any]:
        """Update a group's display name using PATCH"""
        payload = {
            "schemas": ["urn:ietf:params:scim:api:messages:2.0:PatchOp"],
            "Operations": [
                {
                    "op": "replace",
                    "path": "displayName",
                    "value": display_name
                }
            ]
        }

        response = requests.patch(
            f"{self.base_url}/v2/Groups/{group_id}",
            headers=self.headers,
            json=payload
        )
        response.raise_for_status()
        return response.json()

    def delete_group(self, group_id: str) -> None:
        """Delete a SCIM group"""
        response = requests.delete(
            f"{self.base_url}/v2/Groups/{group_id}",
            headers=self.headers
        )
        response.raise_for_status()


def main():
    """
    SCIM Demo Client

    This demo uses email addresses with @example.com domain. Before running:
    1. Ensure your organization's allowed_email_domains setting includes "example.com"
    2. SCIM respects organization email domain restrictions - users with disallowed domains will be rejected
    3. If you prefer different domains, update the email addresses below to match your organization's settings
    """
    base_url = os.getenv('SCIM_BASE_URL')
    auth_token = os.getenv('SCIM_AUTH_TOKEN')

    if not base_url or not auth_token:
        print("Error: SCIM_BASE_URL and SCIM_AUTH_TOKEN environment variables must be set")
        sys.exit(1)

    client = SCIMClient(base_url, auth_token)

    print("=" * 60)
    print("SCIM Client Demo")
    print("=" * 60)
    print("\nNOTE: This demo uses @example.com email addresses.")
    print("Ensure your organization's allowed_email_domains includes example.com")
    print("\nThis demo handles conflicts gracefully and can be run multiple times.")
    print("If resources already exist, they will be reused instead of failing.")

    try:
        print("\n1. Creating a test user...")
        print("   [POST /v2/Users] Schema: urn:ietf:params:scim:schemas:core:2.0:User")
        username = "demo.user@example.com"
        try:
            user = client.create_user(
                username=username,
                given_name="Demo",
                family_name="User",
                active=True
            )
            user_id = user['id']
            print(f"   Created user: {user['userName']} (ID: {user_id})")
            print(f"   Display name: {user['displayName']}")
            print(f"   Active: {user['active']}")
        except requests.exceptions.HTTPError as e:
            if e.response.status_code == 409:
                print(f"   User already exists, finding existing user...")
                print(f"   [GET /v2/Users?filter=userName eq \"{username}\"]")
                user = client.find_user_by_username(username)
                if user:
                    user_id = user['id']
                    print(f"   Found user: {user['userName']} (ID: {user_id})")
                    print(f"   Display name: {user['displayName']}")
                    print(f"   Active: {user['active']}")
                else:
                    raise Exception(f"User exists but could not be found: {username}")
            else:
                raise

        print("\n2. Creating a test group...")
        print("   [POST /v2/Groups] Schema: urn:ietf:params:scim:schemas:core:2.0:Group")
        group_name = "Demo Engineering Team"
        try:
            group = client.create_group(
                display_name=group_name,
                members=[user_id]
            )
            group_id = group['id']
            print(f"   Created group: {group['displayName']} (ID: {group_id})")
            print(f"   Members: {len(group.get('members', []))}")
        except requests.exceptions.HTTPError as e:
            if e.response.status_code == 409:
                print(f"   Group already exists, finding existing group...")
                print(f"   [GET /v2/Groups?filter=displayName eq \"{group_name}\"]")
                group = client.find_group_by_name(group_name)
                if group:
                    group_id = group['id']
                    print(f"   Found group: {group['displayName']} (ID: {group_id})")
                    print(f"   Members: {len(group.get('members', []))}")
                else:
                    raise Exception(f"Group exists but could not be found: {group_name}")
            else:
                raise

        print("\n3. Getting user by ID...")
        print(f"   [GET /v2/Users/{user_id}]")
        fetched_user = client.get_user(user_id)
        print(f"   Retrieved user: {fetched_user['userName']}")
        print(f"   Display name: {fetched_user['displayName']}")

        print("\n4. Getting group by ID...")
        print(f"   [GET /v2/Groups/{group_id}]")
        fetched_group = client.get_group(group_id)
        print(f"   Retrieved group: {fetched_group['displayName']}")
        print(f"   Members: {len(fetched_group.get('members', []))}")

        print("\n5. Listing all users...")
        print("   [GET /v2/Users] With pagination")
        users_list = client.list_users()
        print(f"   Total users: {users_list.get('totalResults', 0)}")

        print("\n6. Listing all groups...")
        print("   [GET /v2/Groups] With pagination")
        groups_list = client.list_groups()
        print(f"   Total groups: {groups_list.get('totalResults', 0)}")

        print("\n7. Creating a second user...")
        print("   [POST /v2/Users] Schema: urn:ietf:params:scim:schemas:core:2.0:User")
        username2 = "jane.doe@example.com"
        try:
            user2 = client.create_user(
                username=username2,
                given_name="Jane",
                family_name="Doe",
                active=True
            )
            user2_id = user2['id']
            print(f"   Created user: {user2['userName']} (ID: {user2_id})")
        except requests.exceptions.HTTPError as e:
            if e.response.status_code == 409:
                print(f"   User already exists, finding existing user...")
                print(f"   [GET /v2/Users?filter=userName eq \"{username2}\"]")
                user2 = client.find_user_by_username(username2)
                if user2:
                    user2_id = user2['id']
                    print(f"   Found user: {user2['userName']} (ID: {user2_id})")
                else:
                    raise Exception(f"User exists but could not be found: {username2}")
            else:
                raise

        print("\n8. Adding second user to group...")
        print(f"   [PATCH /v2/Groups/{group_id}] Schema: urn:ietf:params:scim:api:messages:2.0:PatchOp")
        print("   Operation: add members")
        updated_group = client.add_group_member(group_id, user2_id)
        print(f"   Group now has {len(updated_group.get('members', []))} members")

        print("\n9. Updating group name...")
        print(f"   [PATCH /v2/Groups/{group_id}] Schema: urn:ietf:params:scim:api:messages:2.0:PatchOp")
        print("   Operation: replace displayName")
        new_group_name = "Demo Engineering Team - Updated"
        updated_group = client.update_group_name(group_id, new_group_name)
        print(f"   Updated group name to: {updated_group['displayName']}")

        print("\n10. Updating user name...")
        print(f"   [PATCH /v2/Users/{user_id}] Schema: urn:ietf:params:scim:api:messages:2.0:PatchOp")
        print("   Operation: replace name.givenName, name.familyName")
        updated_user = client.update_user_name(user_id, "Demo", "UserUpdated")
        print(f"   Updated user display name to: {updated_user['displayName']}")

        print("\n11. Removing second user from group...")
        print(f"   [PATCH /v2/Groups/{group_id}] Schema: urn:ietf:params:scim:api:messages:2.0:PatchOp")
        print("   Operation: remove members")
        updated_group = client.remove_group_member(group_id, user2_id)
        print(f"   Group now has {len(updated_group.get('members', []))} members")

        print("\n12. Deactivating the first user...")
        print(f"   [PATCH /v2/Users/{user_id}] Schema: urn:ietf:params:scim:api:messages:2.0:PatchOp")
        print("   Operation: replace active=false")
        updated_user = client.deactivate_user(user_id)
        print(f"   User active status: {updated_user['active']}")

        print("\n13. Reactivating the first user...")
        print(f"   [PATCH /v2/Users/{user_id}] Schema: urn:ietf:params:scim:api:messages:2.0:PatchOp")
        print("   Operation: replace active=true")
        updated_user = client.activate_user(user_id)
        print(f"   User active status: {updated_user['active']}")

        print("\n14. Cleaning up...")
        print(f"   [DELETE /v2/Groups/{group_id}]")
        try:
            client.delete_group(group_id)
            print(f"   Deleted group: {group_id}")
        except requests.exceptions.HTTPError as e:
            if e.response.status_code == 404:
                print(f"   Group already deleted: {group_id}")
            else:
                raise

        print(f"   [DELETE /v2/Users/{user_id}]")
        try:
            client.delete_user(user_id)
            print(f"   Deleted first user: {user_id}")
        except requests.exceptions.HTTPError as e:
            if e.response.status_code == 404:
                print(f"   User already deleted: {user_id}")
            else:
                raise

        print(f"   [DELETE /v2/Users/{user2_id}]")
        try:
            client.delete_user(user2_id)
            print(f"   Deleted second user: {user2_id}")
        except requests.exceptions.HTTPError as e:
            if e.response.status_code == 404:
                print(f"   User already deleted: {user2_id}")
            else:
                raise

        print("\n" + "=" * 60)
        print("Demo completed successfully!")
        print("=" * 60)

    except requests.exceptions.HTTPError as e:
        print(f"\nHTTP Error: {e}")
        if e.response is not None:
            try:
                error_detail = e.response.json()
                print(f"Error details: {json.dumps(error_detail, indent=2)}")
            except:
                print(f"Response: {e.response.text}")
        sys.exit(1)
    except Exception as e:
        print(f"\nError: {e}")
        sys.exit(1)


if __name__ == '__main__':
    main()
