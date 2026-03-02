import 'package:flutter_test/flutter_test.dart';
import 'package:lumber_now/models/models.dart';
import 'package:lumber_now/models/pagination.dart';

void main() {
  group('TenantConfig', () {
    test('fromJson parses all fields', () {
      final json = {
        'dealer_id': 'abc-123',
        'name': 'Lumber Boss',
        'slug': 'lumber-boss',
        'logo_url': 'https://example.com/logo.png',
        'primary_color': '#FF0000',
        'secondary_color': '#00FF00',
        'contact_email': 'info@lumberboss.com',
        'contact_phone': '555-1234',
      };

      final config = TenantConfig.fromJson(json);

      expect(config.dealerId, 'abc-123');
      expect(config.name, 'Lumber Boss');
      expect(config.slug, 'lumber-boss');
      expect(config.logoUrl, 'https://example.com/logo.png');
      expect(config.primaryColor, '#FF0000');
      expect(config.secondaryColor, '#00FF00');
      expect(config.contactEmail, 'info@lumberboss.com');
      expect(config.contactPhone, '555-1234');
    });

    test('fromJson uses defaults for missing fields', () {
      final config = TenantConfig.fromJson({});

      expect(config.dealerId, '');
      expect(config.name, '');
      expect(config.primaryColor, '#1E40AF');
      expect(config.secondaryColor, '#1E3A5F');
    });
  });

  group('User', () {
    test('fromJson parses all fields', () {
      final json = {
        'id': 'user-1',
        'dealer_id': 'dealer-1',
        'email': 'test@example.com',
        'full_name': 'Test User',
        'phone': '555-0000',
        'role': 'contractor',
        'active': true,
      };

      final user = User.fromJson(json);

      expect(user.id, 'user-1');
      expect(user.dealerId, 'dealer-1');
      expect(user.email, 'test@example.com');
      expect(user.fullName, 'Test User');
      expect(user.role, 'contractor');
      expect(user.active, true);
    });

    test('fromJson uses defaults for missing fields', () {
      final user = User.fromJson({});

      expect(user.id, '');
      expect(user.role, 'contractor');
      expect(user.active, true);
    });
  });

  group('StructuredItem', () {
    test('fromJson parses all fields', () {
      final json = {
        'sku': 'LUM-2x4-8',
        'name': '2x4x8 Lumber',
        'quantity': 50.0,
        'unit': 'EA',
        'confidence': 0.95,
        'matched': true,
        'notes': 'Premium grade',
      };

      final item = StructuredItem.fromJson(json);

      expect(item.sku, 'LUM-2x4-8');
      expect(item.name, '2x4x8 Lumber');
      expect(item.quantity, 50.0);
      expect(item.unit, 'EA');
      expect(item.confidence, 0.95);
      expect(item.matched, true);
      expect(item.notes, 'Premium grade');
    });

    test('fromJson uses defaults for missing fields', () {
      final item = StructuredItem.fromJson({});

      expect(item.sku, '');
      expect(item.quantity, 0.0);
      expect(item.unit, 'EA');
      expect(item.confidence, 0.0);
      expect(item.matched, false);
      expect(item.notes, '');
    });

    test('toJson roundtrips correctly', () {
      final item = StructuredItem(
        sku: 'SKU-1',
        name: 'Test Item',
        quantity: 10.0,
        unit: 'BDL',
        confidence: 0.85,
        matched: true,
        notes: 'Note',
      );

      final json = item.toJson();
      final restored = StructuredItem.fromJson(json);

      expect(restored.sku, item.sku);
      expect(restored.name, item.name);
      expect(restored.quantity, item.quantity);
      expect(restored.unit, item.unit);
      expect(restored.confidence, item.confidence);
      expect(restored.matched, item.matched);
      expect(restored.notes, item.notes);
    });

    test('copyWith creates modified copy', () {
      final item = StructuredItem(
        sku: 'SKU-1',
        name: 'Original',
        quantity: 10.0,
        unit: 'EA',
        confidence: 0.5,
        matched: false,
      );

      final modified = item.copyWith(name: 'Modified', quantity: 20.0, matched: true);

      expect(modified.name, 'Modified');
      expect(modified.quantity, 20.0);
      expect(modified.matched, true);
      // Unchanged fields preserved
      expect(modified.sku, 'SKU-1');
      expect(modified.unit, 'EA');
      expect(modified.confidence, 0.5);
    });
  });

  group('MaterialRequest', () {
    test('fromJson parses all fields including structured items', () {
      final json = {
        'id': 'req-1',
        'dealer_id': 'dealer-1',
        'contractor_id': 'contractor-1',
        'assigned_rep_id': 'rep-1',
        'status': 'parsed',
        'input_type': 'text',
        'raw_text': 'I need 50 2x4s',
        'media_url': '',
        'structured_items': [
          {
            'sku': 'LUM-2x4',
            'name': '2x4 Lumber',
            'quantity': 50,
            'unit': 'EA',
            'confidence': 0.9,
            'matched': true,
          }
        ],
        'ai_confidence': '0.9',
        'notes': '',
        'created_at': '2025-01-01T00:00:00Z',
      };

      final req = MaterialRequest.fromJson(json);

      expect(req.id, 'req-1');
      expect(req.assignedRepId, 'rep-1');
      expect(req.status, 'parsed');
      expect(req.structuredItems.length, 1);
      expect(req.structuredItems[0].sku, 'LUM-2x4');
      expect(req.structuredItems[0].quantity, 50.0);
    });

    test('fromJson handles null structured_items', () {
      final req = MaterialRequest.fromJson({
        'id': 'req-2',
        'structured_items': null,
      });

      expect(req.structuredItems, isEmpty);
    });

    test('fromJson uses defaults for missing fields', () {
      final req = MaterialRequest.fromJson({});

      expect(req.id, '');
      expect(req.status, 'pending');
      expect(req.inputType, 'text');
      expect(req.assignedRepId, isNull);
    });
  });

  group('TokenPair', () {
    test('fromJson parses all fields', () {
      final json = {
        'access_token': 'access-abc',
        'refresh_token': 'refresh-xyz',
        'expires_at': 1700000000,
      };

      final tokens = TokenPair.fromJson(json);

      expect(tokens.accessToken, 'access-abc');
      expect(tokens.refreshToken, 'refresh-xyz');
      expect(tokens.expiresAt, 1700000000);
    });

    test('fromJson uses defaults for missing fields', () {
      final tokens = TokenPair.fromJson({});

      expect(tokens.accessToken, '');
      expect(tokens.refreshToken, '');
      expect(tokens.expiresAt, 0);
    });
  });

  group('PaginatedResponse', () {
    test('hasMore returns true when more items exist', () {
      final response = PaginatedResponse<String>(
        items: ['a', 'b', 'c'],
        total: 10,
        offset: 0,
        limit: 3,
      );

      expect(response.hasMore, true);
      expect(response.nextOffset, 3);
    });

    test('hasMore returns false when all items loaded', () {
      final response = PaginatedResponse<String>(
        items: ['a', 'b'],
        total: 5,
        offset: 3,
        limit: 3,
      );

      expect(response.hasMore, false);
      expect(response.nextOffset, 5);
    });

    test('hasMore returns false for empty response', () {
      final response = PaginatedResponse<String>(
        items: [],
        total: 0,
        offset: 0,
        limit: 50,
      );

      expect(response.hasMore, false);
      expect(response.nextOffset, 0);
    });
  });
}
