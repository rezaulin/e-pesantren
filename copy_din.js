const fs = require('fs');
let code = fs.readFileSync('public/app-extra.js', 'utf8');

// Function 1: loadJadwalSekolahMaster
let start1 = code.indexOf('async function loadJadwalSekolahMaster');
let end1 = code.indexOf('async function openJadwalSekolah');
let f1 = code.substring(start1, end1);

// Function 2: openJadwalSekolah
let start2 = end1;
let end2 = code.indexOf('const oldTambahJadwalSekolah', start2); // wait, it was followed by oldTambahJadwalSekolah
let f2 = code.substring(start2, end2);

// Function 3: tambahJadwalSekolah (the real one)
let start3 = code.indexOf('async function tambahJadwalSekolah');
let end3 = code.indexOf('async function hapusJadwalSekolah', start3);
let f3 = code.substring(start3, end3);

// Function 4: hapusJadwalSekolah (the real one)
let start4 = end3;
let end4 = code.indexOf('async function tambahMapelSekolah', start4);
let f4 = code.substring(start4, end4);

// Additional overrides for tambah and hapus to fix openJadwalSekolah logic
let overrides = `
const oldTambahJadwalDiniyyah = tambahJadwalDiniyyah;
tambahJadwalDiniyyah = async function (kelasId, kelasNama) {
  await oldTambahJadwalDiniyyah(kelasId, kelasNama);
  openJadwalDiniyyah(kelasId, kelasNama);
};

const oldHapusJadwalDiniyyah = hapusJadwalDiniyyah;
hapusJadwalDiniyyah = async function (id, kelasId, kelasNama) {
  await oldHapusJadwalDiniyyah(id, kelasId, kelasNama);
  openJadwalDiniyyah(kelasId, kelasNama);
};
`;

let combined = f1 + '\n' + f2 + '\n' + f3 + '\n' + f4 + '\n';
let din = combined.replace(/Sekolah/g, 'Diniyyah').replace(/sekolah/g, 'diniyyah').replace(/jsMapel/g, 'jdMapel').replace(/jsUstadz/g, 'jdUstadz').replace(/jsHariCb/g, 'jdHariCb').replace(/jsMulai/g, 'jdMulai').replace(/jsSelesai/g, 'jdSelesai');

din += '\n' + overrides;

fs.appendFileSync('public/app-extra.js', '\n// Auto-generated Diniyyah functions\n' + din + '\n');
console.log('Success extracting cleanly!');
