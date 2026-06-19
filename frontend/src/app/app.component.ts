import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';

@Component({
  selector: 'app-root',
  standalone: true,
  imports: [CommonModule],
  templateUrl: './app.component.html',
  styleUrls: ['./app.component.css']
})
export class AppComponent implements OnInit {
  title = 'EEIP - Executive Email Intelligence Platform';
  importantEmails: any[] = [];
  risks: any[] = [];
  activeTab = 'dashboard';

  ngOnInit() {
    // Mock data for the UI since the backend might not have real data yet.
    // In production, this would be an HTTP call to the backend API.
    this.importantEmails = [
      { sender: 'ceo@clientabc.com', subject: 'Urgent: Approval needed for Q3 Contract', priority: 'Critical', category: 'Contrato', aiTone: 'Preocupado' },
      { sender: 'ops@vendor.com', subject: 'Server Outage Alert', priority: 'High', category: 'Incidente', aiTone: 'Neutral' }
    ];
    this.risks = [
      { client: 'Client ABC', riskLevel: 'High', description: 'Pending approval delaying deployment', score: 88 }
    ];
    // Adding dummy data for the rest of tabs to fulfill frontend requirements
    this.inbox = [
      { date: '2026-06-19 10:30', sender: 'john@startup.io', client: 'Startup IO', category: 'Soporte', priority: 'Medium', actionRequired: false, status: 'Unread' },
      { date: '2026-06-19 09:15', sender: 'legal@corp.com', client: 'Corp LLC', category: 'Contrato', priority: 'High', actionRequired: true, status: 'Actioned' }
    ];
    this.commitments = [
      { id: 1, emailSubject: 'Project Timeline Update', description: 'Send revised architecture doc', responsible: 'Me', deadline: '2026-06-21', status: 'Pending' }
    ];
    this.contacts = [
      { name: 'Sarah Connor', email: 'sarah@skynet-resistance.com', company: 'Resistance', type: 'Client', sentimentAvg: 'Frustrado' }
    ];
  }

  inbox: any[] = [];
  commitments: any[] = [];
  contacts: any[] = [];

  setTab(tab: string) {
    this.activeTab = tab;
  }
}
