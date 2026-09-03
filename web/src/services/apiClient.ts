import { API_URL } from "@/config/constants";

export class ApiError extends Error {
  status: number;
  data: any;

  constructor(message: string, status: number, data?: any) {
    super(message);
    this.name = "ApiError";
    this.status = status;
    this.data = data;
  }
}

interface RequestOptions extends RequestInit {
  token?: string | null;
}

export async function apiClient<T>(
  endpoint: string,
  options: RequestOptions = {}
): Promise<T> {
  const { token, headers: customHeaders, ...rest } = options;

  // Retrieve stored access token if none explicitly passed
  let authToken = token;
  if (!authToken && typeof window !== "undefined") {
    authToken = localStorage.getItem("goride_access_token");
  }

  const headers: Record<string, string> = {
    "Content-Type": "application/json",
    ...(customHeaders as Record<string, string>),
  };

  if (authToken) {
    headers["Authorization"] = `Bearer ${authToken}`;
  }

  const url = endpoint.startsWith("http") ? endpoint : `${API_URL}${endpoint}`;

  const response = await fetch(url, {
    headers,
    ...rest,
  });

  if (!response.ok) {
    let errorMessage = `Request failed with status ${response.status}`;
    let errorData = null;

    try {
      const json = await response.json();
      errorData = json;
      if (json && json.error) {
        errorMessage = json.error;
      }
    } catch {
      try {
        const text = await response.text();
        if (text) errorMessage = text;
      } catch {
        // use default message
      }
    }

    throw new ApiError(errorMessage, response.status, errorData);
  }

  // Handle empty 204 responses
  if (response.status === 204) {
    return {} as T;
  }

  const result = await response.json();
  // Support standard API response wrapper { data: T } or direct payload
  return (result && result.data !== undefined ? result.data : result) as T;
}
