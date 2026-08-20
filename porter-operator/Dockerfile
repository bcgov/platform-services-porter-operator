#FROM quay.io/operator-framework/ansible-operator:v1.24.1
FROM registry.redhat.io/openshift4/ose-ansible-rhel9-operator:v4.18

COPY requirements.yml ${HOME}/requirements.yml
RUN ansible-galaxy collection install -r ${HOME}/requirements.yml \
 && chmod -R ug+rwx ${HOME}/.ansible

COPY watches.yaml ${HOME}/watches.yaml
COPY roles/ ${HOME}/roles/
COPY playbooks/ ${HOME}/playbooks/
