export type Role = 'platform_admin' | 'dealer_admin' | 'sales_rep' | 'contractor';
export type RequestStatus = 'pending' | 'processing' | 'parsed' | 'confirmed' | 'sent' | 'failed';
export type InputType = 'text' | 'voice' | 'image' | 'pdf';

export interface Dealer {
  id: string;
  name: string;
  slug: string;
  subdomain: string;
  logo_url: string;
  primary_color: string;
  secondary_color: string;
  contact_email: string;
  contact_phone: string;
  address: string;
  active: boolean;
  created_at: string;
  updated_at: string;
}

export interface User {
  id: string;
  dealer_id: string;
  email: string;
  full_name: string;
  phone: string;
  role: Role;
  assigned_rep_id?: string;
  active: boolean;
  created_at: string;
}

export interface InventoryItem {
  id: string;
  dealer_id: string;
  sku: string;
  name: string;
  description: string;
  category: string;
  unit: string;
  price: string;
  in_stock: boolean;
  metadata: Record<string, unknown>;
  created_at: string;
  updated_at: string;
}

export interface StructuredItem {
  sku: string;
  name: string;
  quantity: number;
  unit: string;
  confidence: number;
  matched: boolean;
  notes?: string;
}

export interface MaterialRequest {
  id: string;
  dealer_id: string;
  contractor_id: string;
  assigned_rep_id?: string;
  status: RequestStatus;
  input_type: InputType;
  raw_text: string;
  media_url: string;
  structured_items: StructuredItem[];
  ai_confidence: string;
  notes: string;
  created_at: string;
  updated_at: string;
}

export interface TenantConfig {
  dealer_id: string;
  name: string;
  slug: string;
  logo_url: string;
  primary_color: string;
  secondary_color: string;
  contact_email: string;
  contact_phone: string;
}

export interface TokenPair {
  access_token: string;
  refresh_token: string;
  expires_at: number;
}
