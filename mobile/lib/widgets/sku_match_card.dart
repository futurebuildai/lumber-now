import 'package:flutter/material.dart';
import '../models/models.dart';

class SKUMatchCard extends StatelessWidget {
  final StructuredItem item;

  const SKUMatchCard({super.key, required this.item});

  @override
  Widget build(BuildContext context) {
    final confidencePercent = (item.confidence * 100).toInt();
    final Color confidenceColor;
    if (item.confidence >= 0.8) {
      confidenceColor = Colors.green;
    } else if (item.confidence >= 0.5) {
      confidenceColor = Colors.orange;
    } else {
      confidenceColor = Colors.red;
    }

    return Semantics(
      label: '${item.name}, ${item.quantity} ${item.unit}, ${confidencePercent}% confidence${item.matched ? ", matched" : ""}',
      child: Card(
      margin: const EdgeInsets.only(bottom: 8),
      child: Padding(
        padding: const EdgeInsets.all(12),
        child: Row(
          children: [
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(item.name,
                      style: Theme.of(context)
                          .textTheme
                          .titleSmall
                          ?.copyWith(fontWeight: FontWeight.w600)),
                  const SizedBox(height: 4),
                  Row(
                    children: [
                      if (item.sku.isNotEmpty) ...[
                        Container(
                          padding: const EdgeInsets.symmetric(
                              horizontal: 6, vertical: 2),
                          decoration: BoxDecoration(
                            color: Colors.grey[200],
                            borderRadius: BorderRadius.circular(4),
                          ),
                          child: Text(item.sku,
                              style: TextStyle(
                                  fontSize: 11,
                                  fontFamily: 'monospace',
                                  color: Colors.grey[700])),
                        ),
                        const SizedBox(width: 8),
                      ],
                      Text('${item.quantity} ${item.unit}',
                          style: TextStyle(
                              color: Colors.grey[600], fontSize: 13)),
                    ],
                  ),
                  if (item.notes.isNotEmpty) ...[
                    const SizedBox(height: 4),
                    Text(item.notes,
                        style: TextStyle(
                            fontSize: 12, color: Colors.grey[500],
                            fontStyle: FontStyle.italic)),
                  ],
                ],
              ),
            ),
            Column(
              children: [
                Container(
                  padding:
                      const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
                  decoration: BoxDecoration(
                    color: confidenceColor.withValues(alpha: 0.1),
                    borderRadius: BorderRadius.circular(12),
                  ),
                  child: Text(
                    '$confidencePercent%',
                    style: TextStyle(
                      color: confidenceColor,
                      fontWeight: FontWeight.bold,
                      fontSize: 13,
                    ),
                  ),
                ),
                const SizedBox(height: 4),
                Icon(
                  item.matched ? Icons.check_circle : Icons.help_outline,
                  color: item.matched ? Colors.green : Colors.orange,
                  size: 20,
                ),
              ],
            ),
          ],
        ),
      ),
    ),
    );
  }
}
