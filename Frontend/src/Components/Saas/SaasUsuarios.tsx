import { useEffect, useState } from "react";
import { saasApi, type UserSaaS } from "@/services/saasService";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { UserPlus, Trash2, Mail, Loader2 } from "lucide-react";
import { toast } from "sonner";

export const SaasUsuarios = () => {
  const tenantId = 1;
  const [users, setUsers] = useState<UserSaaS[]>([]);
  const [loading, setLoading] = useState(true);
  const [newEmail, setNewEmail] = useState("");
  const [newRol, setNewRol] = useState<"admin" | "member">("member");
  const [submitting, setSubmitting] = useState(false);

  useEffect(() => {
    let activo = true;
    saasApi
      .getUsers(tenantId)
      .then((res) => {
        if (activo) {
          setUsers(res.items);
          setLoading(false);
        }
      })
      .catch(() => {
        if (activo) {
          toast.error("Error al cargar la lista de usuarios");
          setLoading(false);
        }
      });

    return () => {
      activo = false;
    };
  }, []);

  const handleCreate = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!newEmail.trim()) return;

    try {
      setSubmitting(true);
      await saasApi.createUser(tenantId, {
        email: newEmail.trim(),
        rol: newRol,
      });
      setNewEmail("");
      toast.success("Usuario agregado al tenant");
      const updated = await saasApi.getUsers(tenantId);
      setUsers(updated.items);
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "Error creando usuario");
    } finally {
      setSubmitting(false);
    }
  };

  const handleDelete = async (userId: number) => {
    if (!confirm("¿Seguro que deseas eliminar este usuario del tenant?")) return;
    try {
      await saasApi.deleteUser(tenantId, userId);
      toast.success("Usuario eliminado");
      setUsers((prev) => prev.filter((u) => u.id !== userId));
    } catch {
      toast.error("Error al eliminar usuario");
    }
  };

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold tracking-tight">Equipo y Usuarios</h1>
        <p className="text-muted-foreground">
          Control de acceso y asignación de roles dentro de la organización.
        </p>
      </div>

      {/* Formulario Agregar Usuario */}
      <Card>
        <CardHeader>
          <CardTitle className="text-lg">Invitar nuevo miembro</CardTitle>
          <CardDescription>
            Agrega usuarios al tenant especificando su correo y rol asignado.
          </CardDescription>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleCreate} className="flex flex-col sm:flex-row gap-3">
            <div className="relative flex-1">
              <Mail className="absolute left-3 top-3 h-4 w-4 text-muted-foreground" />
              <Input
                type="email"
                placeholder="colaborador@empresa.com"
                value={newEmail}
                onChange={(e) => setNewEmail(e.target.value)}
                className="pl-9"
                required
              />
            </div>
            <select
              value={newRol}
              onChange={(e) => setNewRol(e.target.value as "admin" | "member")}
              className="h-10 px-3 rounded-md border border-input bg-background text-sm"
            >
              <option value="member">Miembro (member)</option>
              <option value="admin">Administrador (admin)</option>
            </select>
            <Button type="submit" disabled={submitting}>
              {submitting ? <Loader2 className="h-4 w-4 mr-2 animate-spin" /> : <UserPlus className="h-4 w-4 mr-2" />}
              Agregar
            </Button>
          </form>
        </CardContent>
      </Card>

      {/* Tabla de Usuarios */}
      <Card>
        <CardHeader>
          <CardTitle className="text-lg">Usuarios Activos ({users.length})</CardTitle>
        </CardHeader>
        <CardContent>
          {loading ? (
            <div className="text-center py-8 text-sm text-muted-foreground">Cargando usuarios...</div>
          ) : users.length === 0 ? (
            <div className="text-center py-8 text-sm text-muted-foreground">
              No hay usuarios adicionales registrados en este tenant.
            </div>
          ) : (
            <div className="rounded-md border">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Email</TableHead>
                    <TableHead>Rol SaaS</TableHead>
                    <TableHead>Estado</TableHead>
                    <TableHead>Fecha Creación</TableHead>
                    <TableHead className="text-right">Acciones</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {users.map((u) => (
                    <TableRow key={u.id}>
                      <TableCell className="font-medium">{u.email}</TableCell>
                      <TableCell>
                        <Badge variant={u.rol === "admin" || u.rol === "superadmin" ? "default" : "secondary"}>
                          {u.rol}
                        </Badge>
                      </TableCell>
                      <TableCell>
                        <span className="inline-flex items-center gap-1.5 text-xs">
                          <span
                            className={`h-2 w-2 rounded-full ${
                              u.is_active ? "bg-emerald-500" : "bg-muted-foreground"
                            }`}
                          />
                          {u.is_active ? "Activo" : "Inactivo"}
                        </span>
                      </TableCell>
                      <TableCell className="text-xs text-muted-foreground">
                        {new Date(u.created_at).toLocaleDateString()}
                      </TableCell>
                      <TableCell className="text-right">
                        <Button
                          variant="ghost"
                          size="icon"
                          className="h-8 w-8 text-muted-foreground hover:text-destructive"
                          onClick={() => handleDelete(u.id)}
                        >
                          <Trash2 className="h-4 w-4" />
                        </Button>
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
};
