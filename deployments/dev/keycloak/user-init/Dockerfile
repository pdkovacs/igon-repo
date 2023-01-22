FROM ubuntu:22.04

LABEL Name="keycloak-init-tf-job"
LABEL Version="1.0.0"

RUN apt-get update && apt-get install -y gnupg software-properties-common wget grep curl jq && apt-get clean -y
RUN apt-key adv --keyserver keyserver.ubuntu.com --recv-keys AA16FCBCA621E701
RUN wget -O- https://apt.releases.hashicorp.com/gpg | gpg --dearmor | tee /usr/share/keyrings/hashicorp-archive-keyring.gpg
RUN gpg --no-default-keyring --keyring /usr/share/keyrings/hashicorp-archive-keyring.gpg --fingerprint | grep "798A EC65 4E5C 1542 8C8E  42EE AA16 FCBC A621 E701"
RUN echo "deb [signed-by=/usr/share/keyrings/hashicorp-archive-keyring.gpg] https://apt.releases.hashicorp.com $(lsb_release -cs) main" | tee /etc/apt/sources.list.d/hashicorp.list
RUN apt update && apt install terraform

RUN useradd -m --shell /bin/false joe

COPY ./wait-for-local-keycloak.sh /home/joe/
COPY ./create-terraform-client.sh /home/joe/
COPY ./config.tf /home/joe/
COPY ./entrypoint.sh /home/joe/
RUN chmod +x /home/joe/entrypoint.sh

USER joe
WORKDIR /home/joe

CMD [ "./entrypoint.sh" ]
