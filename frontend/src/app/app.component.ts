import { Component, OnInit, inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { HttpClient } from '@angular/common/http';
import { FormsModule } from '@angular/forms';

@Component({
  selector: 'app-root',
  standalone: true,
  imports: [CommonModule, FormsModule],
  templateUrl: './app.component.html',
  styleUrls: ['./app.component.css']
})
export class AppComponent implements OnInit {
  title = 'EEIP - Plataforma de Inteligencia Ejecutiva';
  importantEmails: any[] = [];
  inbox: any[] = [];
  risks: any[] = [];
  commitments: any[] = [];
  contacts: any[] = [];
  activeTab = 'dashboard';

  private http = inject(HttpClient);
  private apiUrl = 'http://localhost:10000/api/v1';

  // Configuración de cuenta
  accounts: any[] = [];
  newAccount: any = {
    email_address: '',
    imap_host: '',
    imap_port: 993,
    imap_user: '',
    imap_password: ''
  };
  isSavingAccount = false;
  isTestingConnection = false;
  accountSuccessMessage = '';
  accountErrorMessage = '';
  editingAccountId: string | null = null;

  ngOnInit() {
    this.loadImportantEmails();
    this.loadAccounts();
    this.inbox = [];
    this.risks = [];
    this.commitments = [];
    this.contacts = [];
  }

  loadAccounts() {
    this.http.get<any[]>(`${this.apiUrl}/accounts`).subscribe({
      next: (data) => {
        if (data) {
          this.accounts = data;
        }
      },
      error: (err) => {
        console.error('Error cargando cuentas', err);
      }
    });
  }

  loadImportantEmails() {
    this.http.get<any[]>(`${this.apiUrl}/emails/important?limit=10`).subscribe({
      next: (data) => {
        if (data) {
          this.importantEmails = data;
        }
      },
      error: (err) => {
        console.error('Error cargando correos importantes', err);
      }
    });
  }

  setTab(tab: string) {
    this.activeTab = tab;
  }

  saveAccount() {
    this.isSavingAccount = true;
    this.accountSuccessMessage = '';
    this.accountErrorMessage = '';

    const req = this.editingAccountId 
      ? this.http.put(`${this.apiUrl}/accounts/${this.editingAccountId}`, this.newAccount)
      : this.http.post(`${this.apiUrl}/accounts`, this.newAccount);

    req.subscribe({
      next: (res) => {
        this.isSavingAccount = false;
        this.accountSuccessMessage = this.editingAccountId ? '¡Cuenta actualizada exitosamente!' : '¡Cuenta configurada y guardada exitosamente!';
        this.resetAccountForm();
        this.loadAccounts();
      },
      error: (err) => {
        this.isSavingAccount = false;
        this.accountErrorMessage = 'Error al guardar la cuenta. Revisa la conexión con el servidor.';
      }
    });
  }

  resetAccountForm() {
    this.newAccount = { email_address: '', imap_host: '', imap_port: 993, imap_user: '', imap_password: '' };
    this.editingAccountId = null;
    this.accountSuccessMessage = '';
    this.accountErrorMessage = '';
  }

  editAccount(acc: any) {
    this.editingAccountId = acc.id;
    this.newAccount = { ...acc, imap_password: '' }; // Don't prefill password
    this.accountSuccessMessage = '';
    this.accountErrorMessage = '';
  }

  deleteAccount(id: string) {
    if (confirm('¿Estás seguro de que deseas eliminar esta cuenta y detener su monitoreo?')) {
      this.http.delete(`${this.apiUrl}/accounts/${id}`).subscribe({
        next: () => {
          this.loadAccounts();
        },
        error: (err) => {
          alert('Error al eliminar la cuenta');
        }
      });
    }
  }

  testConnection(acc: any = null) {
    const dataToTest = acc || this.newAccount;
    if (!dataToTest.imap_host || !dataToTest.imap_user || !dataToTest.imap_password) {
      this.accountErrorMessage = 'Completa los campos del servidor, usuario y contraseña para probar la conexión.';
      return;
    }

    this.isTestingConnection = true;
    this.accountSuccessMessage = '';
    this.accountErrorMessage = '';

    this.http.post(`${this.apiUrl}/accounts/test`, dataToTest).subscribe({
      next: () => {
        this.isTestingConnection = false;
        this.accountSuccessMessage = '¡Conexión IMAP exitosa! Las credenciales son válidas.';
      },
      error: (err) => {
        this.isTestingConnection = false;
        const msg = err.error?.details || 'Credenciales inválidas o servidor inalcanzable.';
        this.accountErrorMessage = `Error IMAP: ${msg}`;
      }
    });
  }
}
