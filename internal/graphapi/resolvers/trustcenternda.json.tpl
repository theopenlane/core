{
    "$schema": "http://json-schema.org/draft-07/schema#",
    "type": "object",
    "properties": {
      "pdf_file_id": {
        "type": "string",
        "title": "NDA PDF File Reference",
        "description": "Reference to the uploaded NDA PDF",
        "const": "{{.NDAFileID}}",
        "x-hidden": true
      },
      "trust_center_id": {
        "const": "{{.TrustCenterID}}",
        "title": "Trust Center Reference",
        "description": "Reference to the trust center",
        "x-hidden": true
      },
      "signatory_info": {
        "type": "object",
        "properties": {
          "email": {
            "type": "string",
            "title": "Email Address",
            "format": "email"
          }
        },
        "required": ["email"]
      },
      "acknowledgment": {
        "type": "boolean",
        "title": "I have read and agree to the terms of this Non-Disclosure Agreement"
      },
      "signature_metadata": {
        "type": "object",
        "x-hidden": true,
        "properties": {
          "ip_address": {
            "type": "string"
          },
          "user_agent": {
            "type": "string"
          },
          "timestamp": {
            "type": "string",
            "format": "date-time"
          },
          "pdf_hash": {
            "type": "string",
            "title": "md5 hash of the signed PDF for integrity verification"
          },
          "user_id": {
            "type": "string",
            "title": "User ID"
          }
        },
        "required": ["ip_address", "timestamp", "pdf_hash", "user_id"]
      }
    },
    "required": ["pdf_file_id", "trust_center_id", "signatory_info", "acknowledgment", "signature_metadata"]
  }