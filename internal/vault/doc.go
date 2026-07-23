// Package vault implements a file storage vault with quota management for the
// simplex-node network. It provides file upload/download/delete operations, a sparse
// file reserve for guaranteed baseline storage (16 GB default), peer cache support
// for P2P radio seeding, and directory size tracking. Key types include Service with
// Upload, Download, Delete, SaveNote, List, UsedMB, UsedPct, and FileCount methods.
package vault
