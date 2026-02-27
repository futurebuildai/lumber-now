class TenantConfig {
  final String dealerId;
  final String name;
  final String slug;
  final String logoUrl;
  final String primaryColor;
  final String secondaryColor;
  final String contactEmail;
  final String contactPhone;

  TenantConfig({
    required this.dealerId,
    required this.name,
    required this.slug,
    required this.logoUrl,
    required this.primaryColor,
    required this.secondaryColor,
    required this.contactEmail,
    required this.contactPhone,
  });

  factory TenantConfig.fromJson(Map<String, dynamic> json) => TenantConfig(
        dealerId: json['dealer_id'] ?? '',
        name: json['name'] ?? '',
        slug: json['slug'] ?? '',
        logoUrl: json['logo_url'] ?? '',
        primaryColor: json['primary_color'] ?? '#1E40AF',
        secondaryColor: json['secondary_color'] ?? '#1E3A5F',
        contactEmail: json['contact_email'] ?? '',
        contactPhone: json['contact_phone'] ?? '',
      );
}

class User {
  final String id;
  final String dealerId;
  final String email;
  final String fullName;
  final String phone;
  final String role;
  final bool active;

  User({
    required this.id,
    required this.dealerId,
    required this.email,
    required this.fullName,
    required this.phone,
    required this.role,
    required this.active,
  });

  factory User.fromJson(Map<String, dynamic> json) => User(
        id: json['id'] ?? '',
        dealerId: json['dealer_id'] ?? '',
        email: json['email'] ?? '',
        fullName: json['full_name'] ?? '',
        phone: json['phone'] ?? '',
        role: json['role'] ?? 'contractor',
        active: json['active'] ?? true,
      );
}

class StructuredItem {
  final String sku;
  final String name;
  final double quantity;
  final String unit;
  final double confidence;
  final bool matched;
  final String notes;

  StructuredItem({
    required this.sku,
    required this.name,
    required this.quantity,
    required this.unit,
    required this.confidence,
    required this.matched,
    this.notes = '',
  });

  factory StructuredItem.fromJson(Map<String, dynamic> json) => StructuredItem(
        sku: json['sku'] ?? '',
        name: json['name'] ?? '',
        quantity: (json['quantity'] ?? 0).toDouble(),
        unit: json['unit'] ?? 'EA',
        confidence: (json['confidence'] ?? 0).toDouble(),
        matched: json['matched'] ?? false,
        notes: json['notes'] ?? '',
      );

  Map<String, dynamic> toJson() => {
        'sku': sku,
        'name': name,
        'quantity': quantity,
        'unit': unit,
        'confidence': confidence,
        'matched': matched,
        'notes': notes,
      };

  StructuredItem copyWith({
    String? sku,
    String? name,
    double? quantity,
    String? unit,
    double? confidence,
    bool? matched,
    String? notes,
  }) {
    return StructuredItem(
      sku: sku ?? this.sku,
      name: name ?? this.name,
      quantity: quantity ?? this.quantity,
      unit: unit ?? this.unit,
      confidence: confidence ?? this.confidence,
      matched: matched ?? this.matched,
      notes: notes ?? this.notes,
    );
  }
}

class MaterialRequest {
  final String id;
  final String dealerId;
  final String contractorId;
  final String? assignedRepId;
  final String status;
  final String inputType;
  final String rawText;
  final String mediaUrl;
  final List<StructuredItem> structuredItems;
  final String aiConfidence;
  final String notes;
  final String createdAt;

  MaterialRequest({
    required this.id,
    required this.dealerId,
    required this.contractorId,
    this.assignedRepId,
    required this.status,
    required this.inputType,
    required this.rawText,
    required this.mediaUrl,
    required this.structuredItems,
    required this.aiConfidence,
    required this.notes,
    required this.createdAt,
  });

  factory MaterialRequest.fromJson(Map<String, dynamic> json) {
    final items = (json['structured_items'] as List<dynamic>?)
            ?.map((e) => StructuredItem.fromJson(e as Map<String, dynamic>))
            .toList() ??
        [];
    return MaterialRequest(
      id: json['id'] ?? '',
      dealerId: json['dealer_id'] ?? '',
      contractorId: json['contractor_id'] ?? '',
      assignedRepId: json['assigned_rep_id'],
      status: json['status'] ?? 'pending',
      inputType: json['input_type'] ?? 'text',
      rawText: json['raw_text'] ?? '',
      mediaUrl: json['media_url'] ?? '',
      structuredItems: items,
      aiConfidence: json['ai_confidence']?.toString() ?? '0',
      notes: json['notes'] ?? '',
      createdAt: json['created_at'] ?? '',
    );
  }
}

class TokenPair {
  final String accessToken;
  final String refreshToken;
  final int expiresAt;

  TokenPair({
    required this.accessToken,
    required this.refreshToken,
    required this.expiresAt,
  });

  factory TokenPair.fromJson(Map<String, dynamic> json) => TokenPair(
        accessToken: json['access_token'] ?? '',
        refreshToken: json['refresh_token'] ?? '',
        expiresAt: json['expires_at'] ?? 0,
      );
}
