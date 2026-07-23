import 'package:flutter/material.dart';

/// ContactInfo manages data model for a single chat contact with display name and message count.
class ContactInfo {
  final int id;
  final String displayName;
  final String? fullName;
  final int msgCount;
  ContactInfo({required this.id, required this.displayName, this.fullName, this.msgCount = 0});
}

/// ChatContactList manages a scrollable list of contacts with group filtering, trust badges and unread counts.
class ChatContactList extends StatelessWidget {
  final List<ContactInfo> contacts;
  final List<ContactGroup> groups;
  final String selectedGroup;
  final String? selectedContactId;
  final Map<String, String> contactTrust;
  final Map<String, String> contactStatuses;
  final Map<String, int> perContactUnread;
  final ValueChanged<String>? onContactTap;
  final ValueChanged<ContactInfo>? onContactLongPress;
  final ValueChanged<String>? onGroupChanged;

  const ChatContactList({
    super.key,
    required this.contacts,
    this.groups = const [],
    this.selectedGroup = '',
    this.selectedContactId,
    this.contactTrust = const {},
    this.contactStatuses = const {},
    this.perContactUnread = const {},
    this.onContactTap,
    this.onContactLongPress,
    this.onGroupChanged,
  });

  @override
  Widget build(BuildContext context) {
    final filtered = selectedGroup.isEmpty ? contacts : contacts.where((c) {
      final grp = groups.where((g) => g.name == selectedGroup).firstOrNull;
      return grp != null && grp.memberIds.contains(c.id);
    }).toList();

    return Column(children: [
      Container(
        padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 6),
        child: Row(children: [
          Text('Contacts (${contacts.length})', style: TextStyle(fontSize: 18, fontWeight: FontWeight.bold, color: Colors.grey[400])),
          const Spacer(),
          if (groups.isNotEmpty)
            Padding(
              padding: const EdgeInsets.only(right: 4),
              child: DropdownButton<String>(
                value: selectedGroup.isEmpty ? null : selectedGroup,
                hint: const Text('Group', style: TextStyle(fontSize: 14)),
                items: [const DropdownMenuItem(value: '', child: Text('All', style: TextStyle(fontSize: 14))),
                  ...groups.map((g) => DropdownMenuItem(value: g.name, child: Text(g.name, style: const TextStyle(fontSize: 14)))),
                ],
                onChanged: (v) => onGroupChanged?.call(v ?? ''),
              ),
            ),
        ]),
      ),
      Expanded(
        child: filtered.isEmpty
            ? Center(child: Column(mainAxisSize: MainAxisSize.min, children: [
                Icon(Icons.people_outline, size: 64, color: Colors.grey[700]),
                const SizedBox(height: 12),
                Text('No contacts yet', style: TextStyle(fontSize: 18, color: Colors.grey[500])),
                const SizedBox(height: 4),
                Text('Share QR code from the QR tab', style: TextStyle(fontSize: 16, color: Colors.grey[600])),
              ]))
            : ListView.builder(
                itemCount: filtered.length,
                itemBuilder: (ctx, i) {
                  final c = filtered[i];
                  final chatId = '@${c.id}';
                  final trustLevel = contactTrust[chatId];
                  final status = contactStatuses[chatId];
                  return ListTile(
                    dense: true,
                    leading: Stack(children: [
                      CircleAvatar(radius: 18, backgroundColor: Colors.grey.shade700,
                        child: Text(c.displayName[0].toUpperCase(), style: const TextStyle(fontSize: 16, fontWeight: FontWeight.bold))),
                      if (status != null && status == 'online')
                        Positioned(right: 0, bottom: 0, child: Container(width: 10, height: 10,
                          decoration: BoxDecoration(shape: BoxShape.circle, color: Colors.green, border: Border.all(color: Colors.grey.shade900, width: 2)))),
                    ]),
                    title: Row(children: [
                      if (trustLevel != null && trustLevel != 'none')
                        Padding(
                          padding: const EdgeInsets.only(right: 4),
                          child: Tooltip(
                            message: trustLevel,
                            child: Icon(
                              trustLevel == 'verified' ? Icons.verified : (trustLevel == 'trusted' ? Icons.shield : Icons.block),
                              size: 16, color: trustLevel == 'verified' ? Colors.blue : (trustLevel == 'trusted' ? Colors.green : Colors.red),
                            ),
                          ),
                        ),
                      Expanded(child: Text(c.displayName, style: const TextStyle(fontSize: 18))),
                    ]),
                    subtitle: Row(children: [
                      if (status != null && status != 'offline')
                        Text(status, style: TextStyle(fontSize: 14, color: status == 'online' ? Colors.green[400] : Colors.orange[400])),
                      if (status != null && status != 'offline') const SizedBox(width: 6),
                      if (c.fullName != null && c.fullName!.isNotEmpty)
                        Text(c.fullName!, style: TextStyle(fontSize: 14, color: Colors.grey[500])),
                    ]),
                    trailing: Row(mainAxisSize: MainAxisSize.min, children: [
                      if (perContactUnread[chatId] != null && perContactUnread[chatId]! > 0)
                        Container(
                          padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
                          decoration: BoxDecoration(color: Colors.red, borderRadius: BorderRadius.circular(10)),
                          child: Text('${perContactUnread[chatId]}', style: const TextStyle(fontSize: 12, color: Colors.white, fontWeight: FontWeight.bold)),
                        ),
                      if (c.msgCount > 0)
                        Padding(padding: const EdgeInsets.only(left: 6, right: 8),
                          child: Text('${c.msgCount}', style: TextStyle(fontSize: 14, color: Colors.grey[500]))),
                      Text('#${c.id}', style: TextStyle(fontSize: 14, color: Colors.grey[600])),
                    ]),
                    onTap: () => onContactTap?.call(chatId),
                    onLongPress: () => onContactLongPress?.call(c),
                  );
                },
              ),
      ),
    ]);
  }
}

/// ContactGroup manages data model for a named group of chat contacts.
class ContactGroup {
  final String name;
  final List<int> memberIds;
  ContactGroup({required this.name, required this.memberIds});
  factory ContactGroup.fromJson(Map<String, dynamic> json) {
    return ContactGroup(
      name: json['name'] as String? ?? '',
      memberIds: (json['member_ids'] as List?)?.map((e) => (e as num).toInt()).toList() ?? [],
    );
  }
}
