/// MarketItem manages marketitem functionality.
class MarketItem {
  final String id;
  final String name;
  final int priceNg;
  final bool forSale;
  final String? holder;

  const MarketItem({
    required this.id,
    this.name = '',
    this.priceNg = 0,
    this.forSale = false,
    this.holder,
  });

  factory MarketItem.fromJson(Map<String, dynamic> json) {
    return MarketItem(
      id: json['id'] ?? '',
      name: json['name'] ?? json['id'] ?? '',
      priceNg: (json['price_ng'] ?? 0).toInt(),
      forSale: json['for_sale'] ?? false,
      holder: json['holder'],
    );
  }

  String get priceFormatted {
    if (priceNg >= 1000000000) {
      return '${(priceNg / 1000000000).toStringAsFixed(2)} TLR';
    }
    return '$priceNg ng';
  }
}
