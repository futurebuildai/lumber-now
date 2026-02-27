class PaginatedResponse<T> {
  final List<T> items;
  final int total;
  final int offset;
  final int limit;

  const PaginatedResponse({
    required this.items,
    required this.total,
    required this.offset,
    required this.limit,
  });

  bool get hasMore => offset + items.length < total;
  int get nextOffset => offset + items.length;
}
