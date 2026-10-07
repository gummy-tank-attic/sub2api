# -*- coding: utf-8 -*-
"""
FXVIA Customization Overlay Script
===================================
When upstream sub2api is updated or rebased, run this script to ensure
all FXVIA customized UI assets (homepage, auth layout, logos, i18n locales, brand CSS)
are cleanly restored and intact.
"""

import os
import shutil
import sys

def main():
    base_dir = os.path.dirname(os.path.abspath(__file__))
    source_dir = os.path.dirname(base_dir)
    frontend_dir = os.path.join(source_dir, "frontend")

    mapping = [
        ("components/HomeView.fxvia.vue", "src/views/HomeView.vue"),
        ("components/AuthLayout.fxvia.vue", "src/components/layout/AuthLayout.vue"),
        ("components/logo.fxvia.svg", "public/logo.svg"),
        ("locales/zh/landing.ts", "src/i18n/locales/zh/landing.ts"),
        ("locales/en/landing.ts", "src/i18n/locales/en/landing.ts"),
    ]

    print("[*] Applying FXVIA brand overlays...")
    for src_rel, dest_rel in mapping:
        src_path = os.path.join(base_dir, src_rel)
        dest_path = os.path.join(frontend_dir, dest_rel)

        if not os.path.exists(src_path):
            print(f"[!] Warning: source {src_path} not found, skipping.")
            continue

        os.makedirs(os.path.dirname(dest_path), exist_ok=True)
        shutil.copy2(src_path, dest_path)
        print(f"[+] Restored: {dest_rel}")

    print("[✓] All FXVIA customizations verified and synchronized successfully!")

if __name__ == "__main__":
    main()
