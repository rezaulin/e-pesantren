import re

with open('app-core.js', 'r', encoding='utf-8', errors='ignore') as f:
    code = f.read()

old_ins = """      try {
        const insidental = await api("/api/insidental/santri/" + a.id);
        const pendingIns = (insidental||[]).filter(t => t.status !== 'LUNAS');
        if (pendingIns.length > 0) {
           cards += `<div style="border-top:1px dashed var(--border);padding-top:.8rem;margin-top:.5rem">
              <div style="font-size:.82rem;color:var(--amber);font-weight:600;margin-bottom:.5rem"><i class="ri-receipt-line"></i> Tagihan Insidental:</div>
              ${pendingIns.map(t => `<div style="display:flex;justify-content:space-between;align-items:center;background:#fff9c4;padding:.4rem .6rem;border-radius:6px;margin-bottom:.4rem;font-size:.8rem">
                 <span>${t.nama}</span>
                 <strong style="color:var(--red)">${fmtR(t.sisa_tagihan)}</strong>
              </div>`).join('')}
              <button class="btn btn-gold btn-sm" style="margin-top:.5rem" onclick="showInsidentalWaliModal(${a.id}, '${a.nama.replace(/'/g, "\\\\'")}')">Bayar Insidental</button>
           </div>`;
        }
      } catch (err) {}"""

new_ins = """      try {
        const insidental = await api("/api/insidental/santri/" + a.id);
        const pendingIns = (insidental||[]).filter(t => t.status !== 'LUNAS');
        if (pendingIns.length > 0) {
           cards += `<div style="border-top:1px dashed var(--border);padding-top:.8rem;margin-top:.5rem">
              <div style="font-size:.82rem;color:var(--amber);font-weight:600;margin-bottom:.5rem"><i class="ri-receipt-line"></i> Tagihan Insidental:</div>
              ${pendingIns.map(t => `<div style="display:flex;justify-content:space-between;align-items:center;background:#fff9c4;padding:.4rem .6rem;border-radius:6px;margin-bottom:.4rem;font-size:.8rem">
                 <span>${t.nama}</span>
                 <strong style="color:var(--red)">${fmtR(t.sisa_tagihan)}</strong>
              </div>`).join('')}
              <button class="btn btn-gold btn-sm" style="margin-top:.5rem" onclick="showInsidentalWaliModal(${a.id}, '${a.nama.replace(/'/g, "\\\\'")}')">Bayar Insidental</button>
           </div>`;
        } else if (insidental && insidental.length > 0) {
           cards += `<div style="border-top:1px dashed var(--border);padding-top:.8rem;margin-top:.5rem">
              <div style="font-size:.82rem;color:var(--amber);font-weight:600;margin-bottom:.5rem"><i class="ri-receipt-line"></i> Tagihan Insidental:</div>
              <div style="display:flex;justify-content:space-between;align-items:center;background:#f0fdf4;padding:.4rem .6rem;border-radius:6px;margin-bottom:.4rem;font-size:.8rem;color:var(--green)">
                 <span><i class="ri-checkbox-circle-line"></i> Semua tagihan lunas! Cek Riwayat Pembayaran.</span>
              </div>
           </div>`;
        }
      } catch (err) {}"""

code = code.replace(old_ins, new_ins)

old_label = """        const tipeLabel = t.tipe === "spp" ? "SPP" : "Topup Sangu";
        const tipeColor = t.tipe === "spp" ? "var(--green)" : "var(--blue)";"""
new_label = """        const tipeLabel = t.tipe === "spp" ? "SPP" : (t.tipe === "insidental" ? "Insidental" : "Topup Sangu");
        const tipeColor = t.tipe === "spp" ? "var(--green)" : (t.tipe === "insidental" ? "var(--amber)" : "var(--blue)");"""
code = code.replace(old_label, new_label)

old_bulan = """        let bulanInfo = "";
        if (t.bulan_list && t.bulan_list !== "[]") {
          try {
            const bulanArr = JSON.parse(t.bulan_list);
            bulanInfo = bulanArr
              .map((b) => {
                const d = new Date(b + "-01");
                return d.toLocaleDateString("id-ID", {
                  month: "short",
                  year: "2-digit",
                });
              })
              .join(", ");
          } catch (e) {}
        }"""

new_bulan = """        let bulanInfo = "";
        if (t.bulan_list && t.bulan_list !== "[]") {
          try {
            const bulanArr = JSON.parse(t.bulan_list);
            if (t.tipe === "spp") {
                bulanInfo = bulanArr.map((b) => {
                    const d = new Date(b + "-01");
                    return d.toLocaleDateString("id-ID", { month: "short", year: "2-digit" });
                }).join(", ");
            } else if (t.tipe === "insidental") {
                bulanInfo = "ID: " + bulanArr.join(", ");
            }
          } catch (e) {}
        }"""
code = code.replace(old_bulan, new_bulan)

with open('app-core.js', 'w', encoding='utf-8') as f:
    f.write(code)
print("Updated successfully")
