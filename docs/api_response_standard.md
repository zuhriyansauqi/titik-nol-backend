# API Response Standard

This document outlines the standard JSON response format for the Titik Nol Backend API. Consistency in responses ensures a better developer experience and more robust client integrations.

## Success Response

All successful API responses must return a 2xx HTTP status code and follow this structure:

```json
{
  "success": true,
  "message": "Human-readable success message",
  "data": { ... },
  "meta": {
    "pagination": {
      "total_items": 100,
      "total_pages": 10,
      "current_page": 1,
      "page_size": 10
    }
  }
}
```

- **`success`**: Always `true`.
- **`message`**: A brief explanation of the action taken.
- **`data`**: The actual payload (object or array). Can be `null` if no data is returned (e.g., 204 No Content).
- **`meta`**: Optional metadata, such as pagination info.

## Error Response

Errors follow the principles of **RFC 7807 (Problem Details for HTTP APIs)**. They must return a 4xx or 5xx HTTP status code and follow this structure:

```json
{
  "success": false,
  "message": "Brief error summary",
  "error": {
    "type": "https://api.titiknol.com/probs/validation-error",
    "title": "Bad Request",
    "status": 400,
    "detail": "Detailed explanation of what went wrong",
    "instance": "/api/v1/users",
    "errors": [
      {
        "field": "email",
        "message": "must be a valid email address"
      }
    ]
  }
}
```

- **`success`**: Always `false`.
- **`message`**: A high-level error message for display.
- **`error.type`**: A URI reference that identifies the problem type.
- **`error.title`**: A short, human-readable summary of the problem type.
- **`error.status`**: The HTTP status code.
- **`error.detail`**: A human-readable explanation specific to this occurrence of the problem.
- **`error.instance`**: A URI reference that identifies the specific occurrence of the problem (usually the request path).
- **`error.errors`**: Optional list of specific field errors (e.g., for validation).

## Common HTTP Status Codes

| Code | Usage |
|------|-------|
| 200  | Successful request |
| 201  | Resource created successfully |
| 204  | Successful request with no content to return |
| 400  | Invalid request parameters or body |
| 401  | Authentication required or failed |
| 403  | Authenticated but lacks permission |
| 404  | Resource not found |
| 409  | Resource conflict (e.g., duplicate email) |
| 422  | Validation failed |
| 429  | Too many requests (rate limiting) |
| 500  | Unexpected server error |
