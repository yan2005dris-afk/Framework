import zlib
import struct
import os

class CleanAcademicPDFWriter:
    def __init__(self):
        self.objects = []
        self.pages = []
        self.xobjects = {}
        self.current_page_stream = []
        
    def add_object(self, content):
        self.objects.append(content)
        return len(self.objects)

    def add_image(self, name, png_path):
        with open(png_path, 'rb') as f:
            data = f.read()
        assert data[:8] == b'\x89PNG\r\n\x1a\n'
        w, h, depth, color_type = struct.unpack(">IIBB", data[16:26])
        
        pos = 8
        idat = bytearray()
        while pos < len(data):
            length, chunk_type = struct.unpack(">I4s", data[pos:pos+8])
            chunk_data = data[pos+8:pos+8+length]
            if chunk_type == b'IDAT':
                idat.extend(chunk_data)
            pos += 12 + length
        raw = zlib.decompress(idat)
        
        bpp = 4 if color_type == 6 else 3
        stride = w * bpp
        unfiltered = bytearray(w * h * bpp)
        src_pos = 0
        dst_pos = 0
        prev_row = bytearray(stride)
        
        for row in range(h):
            filter_type = raw[src_pos]
            src_pos += 1
            curr_row = bytearray(raw[src_pos:src_pos+stride])
            src_pos += stride
            
            if filter_type == 0:
                pass
            elif filter_type == 1:
                for i in range(bpp, stride):
                    curr_row[i] = (curr_row[i] + curr_row[i - bpp]) & 0xff
            elif filter_type == 2:
                for i in range(stride):
                    curr_row[i] = (curr_row[i] + prev_row[i]) & 0xff
            elif filter_type == 3:
                for i in range(stride):
                    left = curr_row[i - bpp] if i >= bpp else 0
                    up = prev_row[i]
                    curr_row[i] = (curr_row[i] + ((left + up) // 2)) & 0xff
            elif filter_type == 4:
                for i in range(stride):
                    a = curr_row[i - bpp] if i >= bpp else 0
                    b = prev_row[i]
                    c = prev_row[i - bpp] if i >= bpp else 0
                    p = a + b - c
                    pa = abs(p - a)
                    pb = abs(p - b)
                    pc = abs(p - c)
                    pr = a if (pa <= pb and pa <= pc) else (b if pb <= pc else c)
                    curr_row[i] = (curr_row[i] + pr) & 0xff
            else:
                raise ValueError(f"Unknown filter type {filter_type}")
            
            unfiltered[dst_pos:dst_pos+stride] = curr_row
            dst_pos += stride
            prev_row = curr_row
            
        if color_type == 6:  # RGBA
            rgb = bytearray(w * h * 3)
            for i in range(w * h):
                r = unfiltered[i*4]
                g = unfiltered[i*4 + 1]
                b = unfiltered[i*4 + 2]
                a = unfiltered[i*4 + 3]
                inv_a = 255 - a
                rgb[i*3] = (r * a + 255 * inv_a) // 255
                rgb[i*3 + 1] = (g * a + 255 * inv_a) // 255
                rgb[i*3 + 2] = (b * a + 255 * inv_a) // 255
            rgb_data = rgb
        else:
            rgb_data = unfiltered
            
        compressed_rgb = zlib.compress(rgb_data)
        
        self.xobjects[name] = {
            'w': w,
            'h': h,
            'data': compressed_rgb,
            'obj_id': None
        }

    def start_page(self):
        self.current_page_stream = []

    def end_page(self):
        stream_bytes = "\n".join(self.current_page_stream).encode('latin1')
        stream_obj_id = self.add_object({
            'dict': f"<</Length {len(stream_bytes)}>>",
            'stream': stream_bytes
        })
        self.pages.append(stream_obj_id)

    # Drawing helpers
    def set_fill_color(self, r, g, b):
        self.current_page_stream.append(f"{r:.3f} {g:.3f} {b:.3f} rg")

    def set_stroke_color(self, r, g, b):
        self.current_page_stream.append(f"{r:.3f} {g:.3f} {b:.3f} RG")

    def set_line_width(self, w):
        self.current_page_stream.append(f"{w:.2f} w")

    def draw_rect(self, x, y, w, h, fill=True, stroke=True):
        op = "B" if fill and stroke else ("f" if fill else "S")
        self.current_page_stream.append(f"{x:.2f} {y:.2f} {w:.2f} {h:.2f} re {op}")

    def draw_line(self, x1, y1, x2, y2):
        self.current_page_stream.append(f"{x1:.2f} {y1:.2f} m {x2:.2f} {y2:.2f} l S")

    def draw_text(self, text, x, y, font="F1", size=10, r=0.1, g=0.1, b=0.1):
        normalized = (
            text.replace("\u2022", "-")
                .replace("✓", "-")
                .replace("—", "-")
                .replace("–", "-")
                .replace("’", "'")
                .replace("“", "\"")
                .replace("”", "\"")
                .replace("á", "\\341")
                .replace("é", "\\351")
                .replace("í", "\\355")
                .replace("ó", "\\363")
                .replace("ú", "\\372")
                .replace("ñ", "\\361")
                .replace("Á", "\\301")
                .replace("É", "\\311")
                .replace("Í", "\\315")
                .replace("Ó", "\\323")
                .replace("Ú", "\\332")
                .replace("Ñ", "\\321")
        )
        safe_text = normalized.replace("\\", "\\\\").replace("(", "\\(").replace(")", "\\)")
        for oct_c in ["341", "351", "355", "363", "372", "361", "301", "311", "315", "323", "332", "321"]:
            safe_text = safe_text.replace(f"\\\\{oct_c}", f"\\{oct_c}")

        cmds = [
            "BT",
            f"/{font} {size} Tf",
            f"{r:.3f} {g:.3f} {b:.3f} rg",
            f"1 0 0 1 {x:.2f} {y:.2f} Tm",
            f"({safe_text}) Tj",
            "ET"
        ]
        self.current_page_stream.append(" ".join(cmds))

    def draw_centered_text(self, text, y, font="F1", size=10, r=0.1, g=0.1, b=0.1, page_width=595.28):
        factor = 0.58 if "2" in font or "4" in font else 0.50
        text_width = len(text) * size * factor
        x = (page_width - text_width) / 2.0
        self.draw_text(text, x, y, font=font, size=size, r=r, g=g, b=b)

    def draw_image(self, name, x, y, w, h):
        self.current_page_stream.append("q")
        self.current_page_stream.append(f"{w:.2f} 0 0 {h:.2f} {x:.2f} {y:.2f} cm")
        self.current_page_stream.append(f"/{name} Do")
        self.current_page_stream.append("Q")

    def build_pdf(self, output_path):
        font_f1 = self.add_object({'dict': "<</Type/Font/Subtype/Type1/BaseFont/Helvetica/Encoding/WinAnsiEncoding>>"})
        font_f2 = self.add_object({'dict': "<</Type/Font/Subtype/Type1/BaseFont/Helvetica-Bold/Encoding/WinAnsiEncoding>>"})
        font_f3 = self.add_object({'dict': "<</Type/Font/Subtype/Type1/BaseFont/Courier/Encoding/WinAnsiEncoding>>"})
        font_f4 = self.add_object({'dict': "<</Type/Font/Subtype/Type1/BaseFont/Courier-Bold/Encoding/WinAnsiEncoding>>"})
        
        xobj_dict_entries = []
        for name, img in self.xobjects.items():
            img_dict = (
                f"<</Type/XObject/Subtype/Image/Width {img['w']}/Height {img['h']}"
                f"/ColorSpace/DeviceRGB/BitsPerComponent 8/Filter/FlateDecode/Length {len(img['data'])}>>"
            )
            obj_id = self.add_object({'dict': img_dict, 'stream': img['data']})
            img['obj_id'] = obj_id
            xobj_dict_entries.append(f"/{name} {obj_id} 0 R")
            
        xobj_res_str = f"<<{' '.join(xobj_dict_entries)}>>" if xobj_dict_entries else "<<"
        
        resources_dict = (
            f"<</Font <</F1 {font_f1} 0 R /F2 {font_f2} 0 R /F3 {font_f3} 0 R /F4 {font_f4} 0 R>> "
            f"/XObject {xobj_res_str} /ProcSet [/PDF /Text /ImageC]>>"
        )
        resources_id = self.add_object({'dict': resources_dict})
        
        pages_id = len(self.objects) + len(self.pages) + 1
        page_obj_ids = []
        for stream_id in self.pages:
            p_dict = (
                f"<</Type/Page/Parent {pages_id} 0 R/MediaBox [0 0 595.28 841.89]"
                f"/Resources {resources_id} 0 R/Contents {stream_id} 0 R>>"
            )
            p_id = self.add_object({'dict': p_dict})
            page_obj_ids.append(p_id)
            
        kids_str = " ".join([f"{pid} 0 R" for pid in page_obj_ids])
        pages_dict = f"<</Type/Pages/Kids [{kids_str}]/Count {len(page_obj_ids)}>>"
        actual_pages_id = self.add_object({'dict': pages_dict})
        assert actual_pages_id == pages_id
        
        info_dict = (
            "<</Title (Actividad AA1-5: Endpoints Pedidos y Comparativa MVC, MVT, API-First)"
            "/Author (Yandris Rivera Torres)"
            "/Subject (Frameworks 7mo Semestre - SaaS Multi-tenant FastAPI)"
            "/Creator (Antigravity LaTeX Deliverable Engine)>>"
        )
        info_id = self.add_object({'dict': info_dict})
        catalog_dict = f"<</Type/Catalog/Pages {pages_id} 0 R>>"
        catalog_id = self.add_object({'dict': catalog_dict})
        
        offsets = [0]
        out = bytearray(b"%PDF-1.4\n%\xe2\xe3\xcf\xd3\n")
        
        for i, obj in enumerate(self.objects):
            offsets.append(len(out))
            obj_num = i + 1
            out.extend(f"{obj_num} 0 obj\n".encode('latin1'))
            if isinstance(obj, dict):
                out.extend(obj['dict'].encode('latin1'))
                out.extend(b"\n")
                if 'stream' in obj:
                    out.extend(b"stream\n")
                    out.extend(obj['stream'])
                    out.extend(b"\nendstream\n")
            out.extend(b"endobj\n")
            
        startxref = len(out)
        out.extend(f"xref\n0 {len(self.objects) + 1}\n".encode('latin1'))
        out.extend(b"0000000000 65535 f \n")
        for i in range(1, len(self.objects) + 1):
            out.extend(f"{offsets[i]:010d} 00000 n \n".encode('latin1'))
            
        trailer = (
            f"trailer\n<<\n/Size {len(self.objects) + 1}\n/Root {catalog_id} 0 R\n"
            f"/Info {info_id} 0 R\n>>\nstartxref\n{startxref}\n%%EOF\n"
        )
        out.extend(trailer.encode('latin1'))
        
        with open(output_path, "wb") as f:
            f.write(out)
        print(f"Generated Clean Academic PDF: {output_path} ({len(out)} bytes, {len(self.pages)} pages)")

# Build Clean Document
doc = CleanAcademicPDFWriter()
doc.add_image("LOGO_UPSE", "docs/entregables/Actividad5/capturas/logo_upse.png")
doc.add_image("IMG_OPENAPI", "docs/entregables/Actividad5/capturas/openapi_pedidos.png")
doc.add_image("IMG_PYTEST", "docs/entregables/Actividad5/capturas/pytest_pedidos.png")
doc.add_image("IMG_SUITE", "docs/entregables/Actividad5/capturas/pytest_suite_resumen.png")

# =========================================================================
# PÁGINA 1: CARÁTULA SOBRIA Y ACADÉMICA (ESTILO INSTITUCIONAL FORMAL)
# =========================================================================
doc.start_page()

# Marco fino y sobrio en negro (0.8 pt), sin fondos de colores
doc.set_stroke_color(0.15, 0.15, 0.15)
doc.set_line_width(0.75)
doc.draw_rect(36, 36, 523, 770, fill=False, stroke=True)

# Encabezados Institucionales sobrios
doc.draw_centered_text("UNIVERSIDAD ESTATAL PENÍNSULA DE SANTA ELENA", 755, font="F2", size=12, r=0.1, g=0.1, b=0.1)
doc.draw_centered_text("FACULTAD DE SISTEMAS Y TELECOMUNICACIONES", 735, font="F2", size=11, r=0.15, g=0.15, b=0.15)
doc.draw_centered_text("CARRERA DE INGENIERÍA EN SOFTWARE", 717, font="F1", size=10.5, r=0.2, g=0.2, b=0.2)

# Logo UPSE Centrado
doc.draw_image("LOGO_UPSE", 227.6, 510, 140, 158)

# Datos de Asignatura y Trabajo
doc.draw_centered_text("ASIGNATURA:", 470, font="F2", size=10, r=0.1, g=0.1, b=0.1)
doc.draw_centered_text("FRAMEWORK PROGRAMACIÓN WEB", 453, font="F2", size=12, r=0.1, g=0.1, b=0.1)

doc.draw_centered_text("TEMA:", 415, font="F2", size=10, r=0.1, g=0.1, b=0.1)
doc.draw_centered_text("TRABAJO PRÁCTICO EXPERIMENTAL - ACTIVIDAD AA1-5", 398, font="F2", size=11.5, r=0.1, g=0.1, b=0.1)
doc.draw_centered_text("Desarrollo de Endpoints en FastAPI, Validación con Pydantic y Cuadro Comparativo", 382, font="F1", size=9.5, r=0.25, g=0.25, b=0.25)

doc.draw_centered_text("ELABORADO POR:", 340, font="F2", size=10, r=0.1, g=0.1, b=0.1)
doc.draw_centered_text("YANDRIS RIVERA TORRES", 323, font="F2", size=11.5, r=0.1, g=0.1, b=0.1)

doc.draw_centered_text("CURSO Y PARALELO:", 282, font="F2", size=10, r=0.1, g=0.1, b=0.1)
doc.draw_centered_text("SOFTWARE 7/1", 265, font="F1", size=11, r=0.1, g=0.1, b=0.1)

doc.draw_centered_text("DOCENTE:", 225, font="F2", size=10, r=0.1, g=0.1, b=0.1)
doc.draw_centered_text("ING. CARLOS EDUARDO MUÑOZ MENDOZA", 208, font="F1", size=11, r=0.1, g=0.1, b=0.1)

# Separador y Pie
doc.set_stroke_color(0.5, 0.5, 0.5)
doc.set_line_width(0.5)
doc.draw_line(210, 135, 385, 135)

doc.draw_centered_text("LA LIBERTAD - ECUADOR", 112, font="F2", size=10, r=0.15, g=0.15, b=0.15)
doc.draw_centered_text("2026 - 2027", 96, font="F1", size=9, r=0.3, g=0.3, b=0.3)

doc.end_page()


# =========================================================================
# PÁGINA 2: CONTENIDO FORMAL TIPO ARTÍCULO CIENTÍFICO / INFORME TÉCNICO
# =========================================================================
doc.start_page()

# Encabezado formal estándar (solo línea fina negra, sin banners de color)
doc.draw_text("Universidad Estatal Península de Santa Elena", 42, 804, font="F1", size=8.5, r=0.3, g=0.3, b=0.3)
doc.draw_text("Framework Programación Web - Actividad AA1-5", 345, 804, font="F1", size=8.5, r=0.3, g=0.3, b=0.3)
doc.set_stroke_color(0.2, 0.2, 0.2)
doc.set_line_width(0.75)
doc.draw_line(42, 796, 553, 796)

# Título y Metadata del informe
doc.draw_text("Informe Técnico: Módulo de Pedidos y Aislamiento Multi-Tenant", 42, 772, font="F2", size=13, r=0.1, g=0.1, b=0.1)
doc.draw_text("Rama de Trabajo: yandris-rivera/AA1-5   |   Módulo: fastAPI_SaaS   |   Fecha: Noviembre 2026", 42, 757, font="F1", size=8.5, r=0.35, g=0.35, b=0.35)

# Sección 1
doc.draw_text("1. Implementación de Endpoints y Lógica de Negocio", 42, 735, font="F2", size=11, r=0.1, g=0.1, b=0.1)
doc.draw_text("En src/api-saas se desarrolló el recurso Pedido cumpliendo los requerimientos técnicos y de seguridad:", 42, 722, font="F1", size=8.5, r=0.2, g=0.2, b=0.2)

# Endpoints descritos con texto limpio y tipografía monoespaciada
doc.draw_text("a) Endpoint PUT /api/{tenant_id}/pedidos/{pedido_id}/estado", 52, 703, font="F2", size=9, r=0.1, g=0.1, b=0.1)
doc.draw_text("Modifica el estado del pedido siguiendo una máquina de transiciones finitas (pendiente -> preparando ->", 64, 691, font="F1", size=8.5, r=0.2, g=0.2, b=0.2)
doc.draw_text("en_camino -> entregado / cancelado). Cualquier transición inválida responde con 400 Bad Request.", 64, 680, font="F1", size=8.5, r=0.2, g=0.2, b=0.2)

doc.draw_text("b) Endpoint DELETE /api/{tenant_id}/pedidos/{pedido_id}", 52, 663, font="F2", size=9, r=0.1, g=0.1, b=0.1)
doc.draw_text("Elimina el pedido garantizando estricto aislamiento entre tiendas (retorna 204 No Content). Si el pedido", 64, 651, font="F1", size=8.5, r=0.2, g=0.2, b=0.2)
doc.draw_text("pertenece a otra tienda, responde 404 Not Found para prevenir la enumeración de recursos entre inquilinos.", 64, 640, font="F1", size=8.5, r=0.2, g=0.2, b=0.2)

# Sección 2: Validación Pydantic
doc.draw_text("2. Reglas de Validación Declarativa con Pydantic v2", 42, 618, font="F2", size=11, r=0.1, g=0.1, b=0.1)
doc.draw_text("En app/modules/pedidos/schemas.py se implementó la validación estricta en el borde perimetral:", 42, 605, font="F1", size=8.5, r=0.2, g=0.2, b=0.2)

doc.draw_text("- cantidad: int = Field(..., gt=0)  -> Rechaza cualquier valor menor o igual a cero con 422.", 55, 590, font="F1", size=8.5, r=0.15, g=0.15, b=0.15)
doc.draw_text("- total: Decimal = Field(..., gt=0)  -> Rechaza importes menores o iguales a cero con 422.", 55, 577, font="F1", size=8.5, r=0.15, g=0.15, b=0.15)
doc.draw_text("- Base de Datos: La migración 0005_pedidos.py refuerza con CHECK (cantidad > 0 AND total > 0).", 55, 564, font="F1", size=8.5, r=0.15, g=0.15, b=0.15)
doc.draw_text("- Aislamiento: Dependencia get_tenant_db (403), filtro tenant_id (404) y Row-Level Security en PostgreSQL.", 55, 551, font="F1", size=8.5, r=0.15, g=0.15, b=0.15)

# Sección 3: Tabla comparativa formal (estilo booktabs de LaTeX)
doc.draw_text("3. Cuadro Comparativo de Enfoques Arquitectónicos", 42, 528, font="F2", size=11, r=0.1, g=0.1, b=0.1)

table_y = 516
row_h = 23
headers = ["Criterio", "MVC Clásico", "MVT (Django)", "API-First (FastAPI)"]
col_w = [95, 135, 135, 146]

# Línea superior formal gruesa (toprule)
doc.set_stroke_color(0.1, 0.1, 0.1)
doc.set_line_width(1.2)
doc.draw_line(42, table_y, 553, table_y)

# Encabezados de tabla
x_cur = 46
for i, h_txt in enumerate(headers):
    doc.draw_text(h_txt, x_cur, table_y - 12, font="F2", size=8.5, r=0.1, g=0.1, b=0.1)
    x_cur += col_w[i]

# Línea divisoria bajo encabezado (midrule)
doc.set_line_width(0.6)
doc.draw_line(42, table_y - 17, 553, table_y - 17)

rows_data = [
    ("Organización", "Controlador + Vista + Modelo", "Modelo + Template + Vista", "Recurso + Schema + Router"),
    ("Entrada Usuario", "Formulario POST tradicional", "Formulario renderizado servidor", "Payload JSON sobre HTTP"),
    ("Validación", "Manual en controlador/modelo", "forms.Form / ModelForm", "Pydantic v2 en el borde (gt=0)"),
    ("Contrato API", "Implícito acoplado a la vista", "Implícito acoplado a templates", "OpenAPI (Swagger) explícito"),
    ("Formato Respuesta", "Páginas HTML completas", "HTML procesado con DTL", "JSON estructurado desacoplado"),
    ("Multi-tenancy", "Filtro manual por consulta", "Filtro manual en manager", "Inyección deps + RLS PostgreSQL"),
    ("Testabilidad", "Requiere servidor HTTP completo", "Requiere django.test.Client", "ASGITransport in-process veloz"),
    ("Acoplamiento", "Alto entre Vista y Controlador", "Alto entre Template y Modelo", "Bajo: Dominio y contratos libres"),
    ("Caso de Uso", "Sitios monolíticos web", "Portales con CMS / Admin", "SaaS multi-cliente (Web/Móvil/APIs)"),
]

cur_y = table_y - 17
for idx, row in enumerate(rows_data):
    cur_y -= row_h
    # Línea muy sutil entre filas
    doc.set_stroke_color(0.88, 0.88, 0.88)
    doc.set_line_width(0.3)
    doc.draw_line(42, cur_y, 553, cur_y)
    
    x_cur = 46
    doc.draw_text(row[0], x_cur, cur_y + 7, font="F2", size=7.5, r=0.1, g=0.1, b=0.1)
    x_cur += col_w[0]
    doc.draw_text(row[1], x_cur, cur_y + 7, font="F1", size=7.5, r=0.2, g=0.2, b=0.2)
    x_cur += col_w[1]
    doc.draw_text(row[2], x_cur, cur_y + 7, font="F1", size=7.5, r=0.2, g=0.2, b=0.2)
    x_cur += col_w[2]
    doc.draw_text(row[3], x_cur, cur_y + 7, font="F1", size=7.5, r=0.1, g=0.1, b=0.1)

# Línea inferior formal (bottomrule)
doc.set_stroke_color(0.1, 0.1, 0.1)
doc.set_line_width(1.2)
doc.draw_line(42, cur_y, 553, cur_y)

# Discusión técnica
cur_y -= 16
doc.draw_text("Análisis Técnico:", 42, cur_y, font="F2", size=8.5, r=0.1, g=0.1, b=0.1)
doc.draw_text("Mientras que MVC y MVT priorizan el renderizado en servidor, la arquitectura API-First sitúa el contrato", 115, cur_y, font="F1", size=8, r=0.2, g=0.2, b=0.2)
cur_y -= 11
doc.draw_text("OpenAPI como artefacto rector. La validación declarativa Pydantic previene procesamiento erróneo de datos antes", 42, cur_y, font="F1", size=8, r=0.2, g=0.2, b=0.2)
cur_y -= 11
doc.draw_text("de llegar a la base de datos, garantizando un contrato desacoplado y reutilizable por múltiples aplicaciones cliente.", 42, cur_y, font="F1", size=8, r=0.2, g=0.2, b=0.2)

# Footer formal
doc.set_stroke_color(0.4, 0.4, 0.4)
doc.set_line_width(0.5)
doc.draw_line(42, 45, 553, 45)
doc.draw_text("Facultad de Sistemas y Telecomunicaciones  -  Carrera de Ingeniería en Software", 42, 34, font="F1", size=8, r=0.4, g=0.4, b=0.4)
doc.draw_text("Página 2 de 3", 500, 34, font="F1", size=8, r=0.4, g=0.4, b=0.4)

doc.end_page()


# =========================================================================
# PÁGINA 3: EVIDENCIAS EN /DOCS, PRUEBAS Y CONCLUSIÓN FORMAL
# =========================================================================
doc.start_page()

# Encabezado formal estándar
doc.draw_text("Universidad Estatal Península de Santa Elena", 42, 804, font="F1", size=8.5, r=0.3, g=0.3, b=0.3)
doc.draw_text("Framework Programación Web - Actividad AA1-5", 345, 804, font="F1", size=8.5, r=0.3, g=0.3, b=0.3)
doc.set_stroke_color(0.2, 0.2, 0.2)
doc.set_line_width(0.75)
doc.draw_line(42, 796, 553, 796)

# Sección 4: OpenAPI
doc.draw_text("4. Evidencias de Pruebas y Contrato en /docs", 42, 772, font="F2", size=11, r=0.1, g=0.1, b=0.1)
doc.draw_text("La especificación OpenAPI generada por FastAPI documenta las rutas implementadas y las reglas de validación:", 42, 759, font="F1", size=8.5, r=0.25, g=0.25, b=0.25)

# OpenAPI Screenshot con marco sutil gris
doc.set_stroke_color(0.7, 0.7, 0.7)
doc.set_line_width(0.5)
doc.draw_rect(41.5, 629.5, 512, 117, fill=False, stroke=True)
doc.draw_image("IMG_OPENAPI", 42, 630, 511, 116)
doc.draw_text("Figura 1: Especificación interactiva OpenAPI (/docs) con PUT /estado, DELETE, POST y GET.", 42, 617, font="F1", size=7.5, r=0.35, g=0.35, b=0.35)

# Sección 5: Pruebas unitarias y de integración
doc.draw_text("5. Verificación Automatizada (Pytest)", 42, 597, font="F2", size=11, r=0.1, g=0.1, b=0.1)
doc.draw_text("Resultados de ejecución de la suite de pruebas del módulo de pedidos y del microservicio completo:", 42, 584, font="F1", size=8.5, r=0.25, g=0.25, b=0.25)

doc.set_stroke_color(0.7, 0.7, 0.7)
doc.set_line_width(0.5)
doc.draw_rect(41.5, 373.5, 281, 199, fill=False, stroke=True)
doc.draw_image("IMG_PYTEST", 42, 374, 280, 198)

doc.draw_rect(330.5, 373.5, 223, 199, fill=False, stroke=True)
doc.draw_image("IMG_SUITE", 331, 374, 222, 198)

doc.draw_text("Figura 2: 26 pruebas de pedidos (gt=0, RLS, 404/403).", 42, 361, font="F1", size=7.5, r=0.35, g=0.35, b=0.35)
doc.draw_text("Figura 3: Suite completa SaaS (83 tests en verde).", 331, 361, font="F1", size=7.5, r=0.35, g=0.35, b=0.35)

# Sección 6: Conclusiones
doc.draw_text("6. Conclusiones", 42, 338, font="F2", size=11, r=0.1, g=0.1, b=0.1)

# Conclusiones presentadas con viñetas limpias y texto académico
doc.draw_text("- Cumplimiento de Endpoints:", 52, 320, font="F2", size=8.5, r=0.1, g=0.1, b=0.1)
doc.draw_text("Se implementaron PUT /api/{tenant_id}/pedidos/{pedido_id}/estado y DELETE /api/{tenant_id}/pedidos/{pedido_id},", 185, 320, font="F1", size=8, r=0.2, g=0.2, b=0.2)
doc.draw_text("garantizando transiciones controladas de estado y aislamiento estricto por tienda.", 185, 310, font="F1", size=8, r=0.2, g=0.2, b=0.2)

doc.draw_text("- Validación en el Borde:", 52, 292, font="F2", size=8.5, r=0.1, g=0.1, b=0.1)
doc.draw_text("La configuración de Field(gt=0) en Pydantic asegura que solicitudes con cantidad o total <= 0", 185, 292, font="F1", size=8, r=0.2, g=0.2, b=0.2)
doc.draw_text("sean rechazadas de forma perimetral con código 422 Unprocessable Entity.", 185, 282, font="F1", size=8, r=0.2, g=0.2, b=0.2)

doc.draw_text("- Aislamiento Multi-Tenant:", 52, 264, font="F2", size=8.5, r=0.1, g=0.1, b=0.1)
doc.draw_text("Se validó en 3 capas: alcance del token (403), filtro de repositorio (404) y Row-Level Security (RLS) en", 185, 264, font="F1", size=8, r=0.2, g=0.2, b=0.2)
doc.draw_text("PostgreSQL configurado en la sesión de base de datos.", 185, 254, font="F1", size=8, r=0.2, g=0.2, b=0.2)

doc.draw_text("- Cobertura de Pruebas:", 52, 236, font="F2", size=8.5, r=0.1, g=0.1, b=0.1)
doc.draw_text("Las 26 pruebas del módulo de pedidos y las 83 pruebas de la suite SaaS confirman la ausencia de regresiones", 185, 236, font="F1", size=8, r=0.2, g=0.2, b=0.2)
doc.draw_text("y el funcionamiento correcto de todos los flujos del sistema.", 185, 226, font="F1", size=8, r=0.2, g=0.2, b=0.2)

# Enlace de repositorio sobrio
doc.set_stroke_color(0.7, 0.7, 0.7)
doc.set_line_width(0.5)
doc.draw_line(42, 195, 553, 195)
doc.draw_text("Repositorio Git: https://github.com/yan2005dris-afk/Framework", 42, 180, font="F2", size=8, r=0.1, g=0.1, b=0.1)
doc.draw_text("Rama de implementación: yandris-rivera/AA1-5  -  Commit integrado en develop y main.", 42, 168, font="F1", size=8, r=0.3, g=0.3, b=0.3)

# Footer formal
doc.set_stroke_color(0.4, 0.4, 0.4)
doc.set_line_width(0.5)
doc.draw_line(42, 45, 553, 45)
doc.draw_text("Facultad de Sistemas y Telecomunicaciones  -  Carrera de Ingeniería en Software", 42, 34, font="F1", size=8, r=0.4, g=0.4, b=0.4)
doc.draw_text("Página 3 de 3", 500, 34, font="F1", size=8, r=0.4, g=0.4, b=0.4)

doc.end_page()

# Generate Clean Academic PDF
output_pdf_path = "docs/entregables/Actividad5/Actividad_AA1_5_Reporte_Final.pdf"
doc.build_pdf(output_pdf_path)
