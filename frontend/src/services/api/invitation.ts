/**
 * Public Invitation API service (no auth required)
 */

import axios from 'axios';
import type { AcceptInvitationRequest } from '../../types/user';

const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080';
const BASE = `${API_BASE_URL}/api/v1/invitations`;

export interface InvitationValidation {
  email: string;
  display_name: string;
  expires_at: string;
}

export async function validateInvitation(
  token: string
): Promise<InvitationValidation> {
  const { data } = await axios.get(`${BASE}/${encodeURIComponent(token)}/validate`);
  return data;
}

export async function acceptInvitation(
  token: string,
  req: AcceptInvitationRequest
): Promise<void> {
  await axios.post(`${BASE}/${encodeURIComponent(token)}/accept`, req);
}
