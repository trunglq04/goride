import { apiClient } from "./apiClient";
import { BackendEndpoints } from "@/types/contracts";
import {
  HTTPTripPreviewRequestPayload,
  HTTPTripPreviewResponse,
  HTTPTripStartRequestPayload,
  HTTPTripStartResponse,
} from "@/types";

export const tripService = {
  /**
   * Request dynamic route pricing and preview fares for origin & destination
   */
  async previewTrip(
    payload: HTTPTripPreviewRequestPayload,
    token?: string | null
  ): Promise<HTTPTripPreviewResponse> {
    return apiClient<HTTPTripPreviewResponse>(BackendEndpoints.PREVIEW_TRIP, {
      method: "POST",
      body: JSON.stringify(payload),
      token,
    });
  },

  /**
   * Confirm selected vehicle tier and start trip matching
   */
  async startTrip(
    payload: HTTPTripStartRequestPayload,
    token?: string | null
  ): Promise<HTTPTripStartResponse> {
    return apiClient<HTTPTripStartResponse>(BackendEndpoints.START_TRIP, {
      method: "POST",
      body: JSON.stringify(payload),
      token,
    });
  },

  /**
   * Cancel an ongoing or requested trip
   */
  async cancelTrip(
    payload: { tripID: string; userID: string; reason?: string },
    token?: string | null
  ): Promise<void> {
    return apiClient<void>(BackendEndpoints.CANCEL_TRIP, {
      method: "POST",
      body: JSON.stringify(payload),
      token,
    });
  },
};
