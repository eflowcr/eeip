import { Component, OnInit, inject, ChangeDetectorRef } from '@angular/core';
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
  private cdr = inject(ChangeDetectorRef);
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
  
  // Para pruebas en línea en la lista
  accountTestStatus: { [id: string]: { loading: boolean, message: string, error: boolean } } = {};
  accountSyncStatus: { [id: string]: { loading: boolean, message: string, error: boolean } } = {};

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

  getCriticalCount(): number {
    return this.importantEmails.filter(e => e.priority === 'Critical').length;
  }

  loadInbox() {
    this.http.get<any[]>(`${this.apiUrl}/emails/all?limit=50`).subscribe({
      next: (data) => {
        if (data) {
          this.inbox = data;
        }
      },
      error: (err) => {
        console.error('Error cargando bandeja global', err);
      }
    });
  }

  setTab(tab: string) {
    this.activeTab = tab;
    if (tab === 'dashboard') {
      this.loadImportantEmails();
    } else if (tab === 'inbox') {
      this.loadInbox();
    }
  }

  markAsResolved(emailId: string) {
    this.http.put(`${this.apiUrl}/emails/${emailId}/status`, { status: 'Actioned' }).subscribe({
      next: () => {
        // Remove it from the UI immediately to feel fast
        this.importantEmails = this.importantEmails.filter(e => e.id !== emailId);
        this.inbox = this.inbox.filter(e => e.id !== emailId);
      },
      error: (err) => console.error('Error marking as resolved', err)
    });
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
        this.cdr.detectChanges();
      },
      error: (err) => {
        this.isSavingAccount = false;
        this.accountErrorMessage = 'Error al guardar la cuenta. Revisa la conexión con el servidor.';
        this.cdr.detectChanges();
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
        this.cdr.detectChanges();
      },
      error: (err) => {
        this.isTestingConnection = false;
        const msg = err.error?.details || 'Credenciales inválidas o servidor inalcanzable.';
        this.accountErrorMessage = `Error IMAP: ${msg}`;
        this.cdr.detectChanges();
      }
    });
  }

  testExistingAccount(accountId: string) {
    this.accountTestStatus[accountId] = { loading: true, message: 'Probando...', error: false };
    this.cdr.detectChanges();

    this.http.post(`${this.apiUrl}/accounts/${accountId}/test`, {}).subscribe({
      next: () => {
        this.accountTestStatus[accountId] = { loading: false, message: '¡Conexión exitosa!', error: false };
        this.cdr.detectChanges();
        // Clear message after 3 seconds
        setTimeout(() => {
          if (this.accountTestStatus[accountId]) {
            this.accountTestStatus[accountId].message = '';
            this.cdr.detectChanges();
          }
        }, 3000);
      },
      error: (err) => {
        const msg = err.error?.details || 'Error de conexión';
        this.accountTestStatus[accountId] = { loading: false, message: `Error: ${msg}`, error: true };
        this.cdr.detectChanges();
      }
    });
  }

  syncAccount(accountId: string) {
    this.accountSyncStatus[accountId] = { loading: true, message: 'Sincronizando...', error: false };
    this.cdr.detectChanges();

    this.http.post(`${this.apiUrl}/accounts/${accountId}/sync`, {}).subscribe({
      next: () => {
        this.accountSyncStatus[accountId] = { loading: false, message: '¡Sincronización completada!', error: false };
        this.loadImportantEmails();
        this.cdr.detectChanges();
        setTimeout(() => {
          if (this.accountSyncStatus[accountId]) {
            this.accountSyncStatus[accountId].message = '';
            this.cdr.detectChanges();
          }
        }, 3000);
      },
      error: (err) => {
        const msg = err.error?.details || 'Error de sincronización';
        this.accountSyncStatus[accountId] = { loading: false, message: `Error: ${msg}`, error: true };
        this.cdr.detectChanges();
      }
    });
  }
}
